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

	"github.com/hung0913208/go-algorithm/lib/container"
	"github.com/hung0913208/go-algorithm/lib/logs"
)

const (
	NoError int = iota
	ErrorInitContainer
	ErrorInitSentry
	ErrorInitSql
	ErrorInitRedis
	ErrorInitMemcache
	ErrorRegisterError
	ErrorRegisterSql
	ErrorRegisterRedis
	ErrorRegisterMemcache
	ErrorRegisterSpawn
	ErrorRegisterBot
)

var (
	input   string
	outputs []string
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

func UNUSED(x ...interface{}) {}

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

	err = container.Init(nil)
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

	UNUSED(timeoutDb, timeoutModule, logger)
}
