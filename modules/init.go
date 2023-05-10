package modules

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	sentry "github.com/getsentry/sentry-go"
	trace "github.com/uptrace/uptrace-go/uptrace"
	ot "go.opentelemetry.io/otel"

	"github.com/hung0913208/go-algorithm/lib/container"
	"github.com/hung0913208/go-algorithm/lib/db"
	"github.com/hung0913208/go-algorithm/lib/kv"
	"github.com/hung0913208/go-algorithm/lib/logs"
)

const (
	NoError int = iota
	ErrorInitContainer
	ErrorInitSentry
	ErrorInitUptime
	ErrorInitBizfly
	ErrorInitRedis
	ErrorInitMemcache
	ErrorRegisterCrawl
	ErrorRegisterBot
	ErrorRegisterSql
	ErrorRegisterRedis
	ErrorRegisterMemcache
)

var (
	input   string
	outputs []string

	tracer = ot.Tracer("monitor")
)

func RegisterWithRetry(
	maxRetry int,
	timeout time.Duration,
	handler func() (container.Module, error),
) (container.Module, error) {
	var ret container.Module
	var err error

	for i := 0; i < maxRetry; i++ {
		ret, err = handler()
		if err == nil {
			return ret, err
		}

		time.Sleep(timeout)
	}

	return nil, err
}

func Init(
	modules map[string]bool,
) {
	if len(os.Getenv("UPTRACE_DSN")) > 0 {
		trace.ConfigureOpentelemetry(
			trace.WithDSN(os.Getenv("UPTRACE_DSN")),
			trace.WithServiceName("sre"),
			trace.WithServiceVersion("1.0.0"),
		)

		log.Print("Finish configuring uptrace")
	}

	timeouts := []string{"100", "2", "200"}

	if len(os.Getenv("TIMEOUT")) > 0 {
		timeouts = strings.Split(os.Getenv("TIMEOUT"), ",")
	}

	timeoutDb, err := strconv.Atoi(timeouts[0])
	if err != nil {
		timeoutDb = 100
	}

	timeoutModule, err := strconv.Atoi(timeouts[0])
	if err != nil {
		timeoutModule = 200
	}

	outputs = make([]string, 0)

	// @NOTE: configure opentelemetry
	err = container.Init(tracer)
	if err != nil {
		container.Terminate(
			"Can't setup container to store modules",
			ErrorInitContainer,
		)
	}

	if len(os.Getenv("SENTRY_DSN")) > 0 {
		// @NOTE: configure sentry
		err = sentry.Init(sentry.ClientOptions{
			Dsn:              os.Getenv("SENTRY_DSN"),
			Debug:            true,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			container.Terminate(fmt.Sprintf("sentry.Init: %v", err), ErrorInitSentry)
		}

		log.Print("Finish configuring sentry")
	}

	logger := logs.NewLoggerWithStacktrace()

	defer func() {
		err := recover()

		if err != nil {
			if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
				logger.Warnf("%v", err)
			}
		}

		sentry.Flush(2 * time.Second)
	}()

	// @NOTE: configure error module
	err = container.RegisterSimpleModule(
		"error",
		logs.NewErrorModule(),
		timeoutDb,
	)

	if err != nil {
		container.Terminate("Can't register module `error`", ErrorRegisterSql)
	}

	log.Print("Finish configuring error")

	// @NOTE: configure elephansql database
	if enabled, ok := modules["elephansql"]; ok && enabled {
		port, err := strconv.Atoi(os.Getenv("ELEPHANSQL_PORT"))
		if err != nil {
			container.Terminate("Can't register module `elephansql`", ErrorRegisterSql)
		}

		elephansql, err := db.NewPgModule(
			os.Getenv("ELEPHANSQL_HOST"),
			port,
			os.Getenv("ELEPHANSQL_USERNAME"),
			os.Getenv("ELEPHANSQL_PASSWORD"),
			os.Getenv("ELEPHANSQL_DATABASE"),
		)
		if err != nil {
			container.Terminate("Can't register module `elephansql`", ErrorRegisterSql)
		}

		err = container.RegisterSimpleModule(
			"elephansql",
			elephansql,
			timeoutDb,
		)
		if err != nil {
			container.Terminate("Can't register module `elephansql`", ErrorRegisterSql)
		}

		log.Print("Load module `elephansql` successfully")
	}

	// @NOTE: configure supabase database
	if enabled, ok := modules["supabase"]; ok && enabled {
		port, err := strconv.Atoi(os.Getenv("SUPABASE_PORT"))
		if err != nil {
			container.Terminate("Can't register module `supabase`", ErrorRegisterSql)
		}

		supabase, err := db.NewPgPoolModule(
			os.Getenv("SUPABASE_HOST"),
			port,
			os.Getenv("SUPABASE_USERNAME"),
			os.Getenv("SUPABASE_PASSWORD"),
			os.Getenv("SUPABASE_DATABASE"),
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `supabase`: %v", err),
				ErrorRegisterSql)
		}

		err = container.RegisterSimpleModule(
			"supabase",
			supabase,
			timeoutDb,
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `supabase`: %v", err),
				ErrorRegisterSql)
		}

		log.Print("Load module `supabase` successfully")
	}

	// @NOTE: configure yugabyte database
	if enabled, ok := modules["yugabyte"]; ok && enabled {
		port, err := strconv.Atoi(os.Getenv("YUGABYTE_PORT"))
		if err != nil {
			container.Terminate("Can't register module `yugabyte`", ErrorRegisterSql)
		}

		yugabyte, err := db.NewPgModuleWithSsl(
			os.Getenv("YUGABYTE_HOST"),
			port,
			os.Getenv("YUGABYTE_USERNAME"),
			os.Getenv("YUGABYTE_PASSWORD"),
			os.Getenv("YUGABYTE_DATABASE"),
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `yugabyte`: %v", err),
				ErrorRegisterSql)
		}

		err = container.RegisterSimpleModule(
			"yugabyte",
			yugabyte,
			timeoutDb,
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `yugabyte`: %v", err),
				ErrorRegisterSql)
		}

		log.Print("Load module `yugabyte` successfully")
	}

	// @NOTE: configure redis module
	if enabled, ok := modules["redislab"]; ok && enabled {
		redis, err := kv.NewRedisModule(
			os.Getenv("REDIS_URI"),
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `redis`: %v", err),
				ErrorInitRedis)
		}

		err = container.RegisterSimpleModule(
			"redislab",
			redis,
			timeoutDb,
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `redis`: %v", err),
				ErrorRegisterRedis)
		}
	}

	if enabled, ok := modules["memcachier"]; ok && enabled {
		memcache, err := kv.NewMemcacheModule(
			os.Getenv("MEMCACHIER_HOST"),
			os.Getenv("MEMCACHIER_USERNAME"),
			os.Getenv("MEMCACHIER_PASSWORD"),
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `memcahe`: %v", err),
				ErrorInitMemcache)
		}

		err = container.RegisterSimpleModule(
			"memcachier",
			memcache,
			timeoutDb,
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `memcache`: %v", err),
				ErrorRegisterMemcache)
		}

		log.Print("Load module `memcachier` successfully")
	}

	if enabled, ok := modules["mysql"]; ok && enabled {
		port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
		if err != nil {
			container.Terminate("Can't register module `mysql`", ErrorRegisterSql)
		}

		mysql, err := db.NewMysqlModule(
			os.Getenv("MYSQL_HOST"),
			port,
			os.Getenv("MYSQL_USERNAME"),
			os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_DATABASE"),
		)

		err = container.RegisterSimpleModule(
			"mysql",
			mysql,
			timeoutDb,
		)
		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `mysql`: %v", err),
				ErrorRegisterMemcache)
		}

		log.Print("Load module `mysql` successfully")
	}

	if enabled, ok := modules["mariadb"]; ok && enabled {
		_, err := RegisterWithRetry(
			30,
			time.Duration(1)*time.Second,
			func() (container.Module, error) {
				port, err := strconv.Atoi(os.Getenv("MARIADB_PORT"))
				if err != nil {
					return nil, err
				}

				mariadb, err := db.NewMysqlModule(
					os.Getenv("MARIADB_HOST"),
					port,
					os.Getenv("MARIADB_USERNAME"),
					os.Getenv("MARIADB_PASSWORD"),
					os.Getenv("MARIADB_DATABASE"),
				)

				err = container.RegisterSimpleModule(
					"mariadb",
					mariadb,
					timeoutDb,
				)
				return mariadb, err
			},
		)

		if err != nil {
			container.Terminate(fmt.Sprintf("Can't register module `mariadb`: %v", err),
				ErrorRegisterMemcache)
		}

		log.Print("Load module `mariadb` successfully")
	}

	if enabled, ok := modules["crawl"]; ok && enabled {
	}
}
