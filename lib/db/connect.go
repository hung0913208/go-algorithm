package db

import (
	"fmt"
	"os"
	"time"

	otel "github.com/uptrace/opentelemetry-go-extra/otelgorm"

	"gorm.io/driver/bigquery"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/hung0913208/go-algorithm/lib/container"
)

type Db interface {
	Establish() *gorm.DB
}

type dbImpl struct {
	dbConn    *gorm.DB
	dialector gorm.Dialector

	host, username, password string
	port                     int
}

type dialectorProxyImpl struct {
	dialector gorm.Dialector
}

var (
	_ gorm.Dialector = &dialectorProxyImpl{}
)

func NewBigQuery(
	projectId, dataset string,
	timeout time.Duration,
) (Db, error) {

	bqConn := bigquery.Open(
		fmt.Sprintf(
			"bigquery://%s/%s",
			projectId,
			dataset,
		),
	)

	if len(os.Getenv("SENTRY_DSN")) > 0 {
		newLogger := newGormSentryLogger(
			logger.Config{
				SlowThreshold:             timeout,
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)

		if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
			newLogger = newGormSentryLogger(
				logger.Config{
					SlowThreshold:             timeout,
					LogLevel:                  logger.Error,
					IgnoreRecordNotFoundError: true,
					Colorful:                  true,
				},
			)
		}

		dbConn, err := gorm.Open(
			&dialectorProxyImpl{
				dialector: bqConn,
			},
			&gorm.Config{
				// @NOTE: configure logger
				Logger: newLogger,
			},
		)

		if err != nil {
			return nil, err
		}

		ret := &dbImpl{
			dbConn:    dbConn,
			dialector: bqConn,
		}
		return ret, ret.setupConnectionPool(1, 5, time.Duration(10)*time.Second)
	} else {
		dbConn, err := gorm.Open(
			&dialectorProxyImpl{
				dialector: bqConn,
			},
			&gorm.Config{},
		)

		if err != nil {
			return nil, err
		}

		ret := &dbImpl{
			dbConn:    dbConn,
			dialector: bqConn,
		}
		return ret, ret.setupConnectionPool(1, 5, time.Duration(10)*time.Second)
	}
}

func NewMysql(
	host string,
	port int,
	username, password, database string,
	timeout time.Duration,
	usePreparedStatement bool,
) (Db, error) {
	newLogger := newGormSentryLogger(
		logger.Config{
			SlowThreshold:             timeout,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
		newLogger = newGormSentryLogger(
			logger.Config{
				SlowThreshold:             timeout,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
	}

	mysqlConn := mysql.Open(
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&"+
				"loc=Local&"+
				"interpolateParams=true",
			username, password,
			host, port,
			database,
		),
	)

	dbConn, err := gorm.Open(
		&dialectorProxyImpl{
			dialector: mysqlConn,
		},
		&gorm.Config{
			// @NOTE: most of the case we don't need transaction since we will
			//        handle everything manually so don't need to
			SkipDefaultTransaction: true,

			// @NOTE: cache prepared statement so the future queries might be
			//        speed up
			PrepareStmt: usePreparedStatement,

			// @NOTE: configure logger
			Logger: newLogger,

			// @NOTE: disable nested transaction to make query more simple and fast
			DisableNestedTransaction: true,
		},
	)

	if err != nil {
		return nil, err
	}

	if len(os.Getenv("UPTRACE_DSN")) > 0 {
		if err := dbConn.Use(otel.NewPlugin()); err != nil {
			return nil, err
		}
	}

	ret := &dbImpl{
		dbConn:    dbConn,
		dialector: mysqlConn,
	}
	return ret, ret.setupConnectionPool(5, 5, time.Duration(60)*time.Second)
}

func NewPg(
	host string,
	port int,
	username, password, database string,
	timeout time.Duration,
	sslMode string,
	usePreparedStatement bool,
) (Db, error) {
	newLogger := newGormSentryLogger(
		logger.Config{
			SlowThreshold:             timeout,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
		newLogger = newGormSentryLogger(
			logger.Config{
				SlowThreshold:             timeout,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
	}

	dsnTemplate := "user=%s password=%s host=%s port=%d dbname=%s sslmode=%s"
	pgConn := postgres.New(
		postgres.Config{
			DSN: fmt.Sprintf(
				dsnTemplate,
				username, password,
				host, port,
				database,
				sslMode,
			),
			PreferSimpleProtocol: !usePreparedStatement,
		},
	)

	dbConn, err := gorm.Open(
		&dialectorProxyImpl{
			dialector: pgConn,
		},
		&gorm.Config{
			// @NOTE: most of the case we don't need transaction since we will
			//        handle everything manually so don't need to
			SkipDefaultTransaction: true,

			// @NOTE: cache prepared statement so the future queries might be
			//        speed up
			PrepareStmt: usePreparedStatement,

			// @NOTE: configure logger
			Logger: newLogger,

			// @NOTE: disable nested transaction to make query more simple and fast
			DisableNestedTransaction: true,
		},
	)

	if err != nil {
		return nil, err
	}

	if err := dbConn.Use(otel.NewPlugin()); err != nil {
		return nil, err
	}

	ret := &dbImpl{
		dbConn:    dbConn,
		dialector: pgConn,
	}
	return ret, ret.setupConnectionPool(1, 5, time.Duration(1)*time.Second)
}

func (self *dbImpl) setupConnectionPool(
	maxIdleConnection, maxOpenConnection int,
	maxLifetime time.Duration,
) error {
	sqlDB, err := self.dbConn.DB()
	if err != nil {
		return err
	}

	if len(os.Getenv("VERCEL_ENV")) == 0 {
		sqlDB.SetMaxIdleConns(maxIdleConnection)
		sqlDB.SetMaxOpenConns(maxOpenConnection)
		sqlDB.SetConnMaxLifetime(maxLifetime)
	}
	return nil
}

func (self *dbImpl) Establish() *gorm.DB {
	if os.Getenv("DEBUG") == "true" {
		return self.dbConn.
			Debug().
			WithContext(container.GetContext())
	} else {
		return self.dbConn.WithContext(container.GetContext())
	}
}

func (self *dialectorProxyImpl) Name() string {
	return self.dialector.Name()
}

func (self *dialectorProxyImpl) Initialize(dbSql *gorm.DB) error {
	return self.dialector.Initialize(dbSql)
}

func (self *dialectorProxyImpl) Migrator(dbSql *gorm.DB) gorm.Migrator {
	return self.dialector.Migrator(dbSql)
}

func (self *dialectorProxyImpl) DataTypeOf(field *schema.Field) string {
	return self.dialector.DataTypeOf(field)
}

func (self *dialectorProxyImpl) DefaultValueOf(
	field *schema.Field,
) clause.Expression {
	return self.dialector.DefaultValueOf(field)
}

func (self *dialectorProxyImpl) BindVarTo(
	writer clause.Writer,
	stmt *gorm.Statement,
	v interface{},
) {
	self.dialector.BindVarTo(writer, stmt, v)
}

func (self *dialectorProxyImpl) QuoteTo(writer clause.Writer, data string) {
	self.dialector.QuoteTo(writer, data)
}

func (self *dialectorProxyImpl) Explain(sql string, vars ...interface{}) string {
	return self.dialector.Explain(sql, vars...)
}
