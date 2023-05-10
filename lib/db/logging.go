package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hung0913208/go-algorithm/lib/logs"

	gormlog "gorm.io/gorm/logger"
)

type dbLogWriterImpl struct {
	logger                    logs.Logger
	logLevel                  gormlog.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

var (
	_ gormlog.Interface = &dbLogWriterImpl{}
)

func newGormSentryLogger(config gormlog.Config) gormlog.Interface {
	if config.Colorful {
		return &dbLogWriterImpl{
			logger:                    logs.NewLoggerWithStacktrace(),
			logLevel:                  config.LogLevel,
			slowThreshold:             config.SlowThreshold,
			ignoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
		}
	} else {
		return &dbLogWriterImpl{
			logger:                    logs.NewLogger(),
			logLevel:                  config.LogLevel,
			slowThreshold:             config.SlowThreshold,
			ignoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
		}
	}
}

func (self *dbLogWriterImpl) LogMode(level gormlog.LogLevel) gormlog.Interface {
	self.logLevel = level
	return self
}

func (self *dbLogWriterImpl) Info(ctx context.Context, format string, data ...interface{}) {
	if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
		fmt.Printf(format, data...)
	}

	if self.logLevel > gormlog.Info {
		self.logger.Infof(format, data...)
	}
}

func (self *dbLogWriterImpl) Warn(ctx context.Context, format string, data ...interface{}) {
	if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
		fmt.Printf(format, data...)
	}

	if self.logLevel > gormlog.Warn {
		self.logger.Warnf(format, data...)
	}
}

func (self *dbLogWriterImpl) Error(ctx context.Context, format string, data ...interface{}) {
	if os.Getenv("VERCEL_ENV") != "production" || os.Getenv("DEBUG") == "true" {
		fmt.Printf(format, data...)
	}

	if self.logLevel > gormlog.Error {
		self.logger.Errorf(format, data...)
	}
}

func (self *dbLogWriterImpl) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	if self.logLevel <= gormlog.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil &&
		self.logLevel >= gormlog.Error &&
		(!errors.Is(err, gormlog.ErrRecordNotFound) || !self.ignoreRecordNotFoundError):
		sql, rows := fc()

		if rows == -1 {
			self.logger.Errorf(
				"%s\n[%.3fms] [rows:%v] %s",
				err,
				float64(elapsed.Nanoseconds())/1e6, "-",
				sql,
			)
		} else {
			self.logger.Errorf(
				"%s\n[%.3fms] [rows:%v] %s",
				err,
				float64(elapsed.Nanoseconds())/1e6,
				rows,
				sql,
			)
		}

	case elapsed > self.slowThreshold &&
		self.slowThreshold != 0 &&
		self.logLevel >= gormlog.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", self.slowThreshold)

		if rows == -1 {
			self.logger.Warnf(
				"%s\n[%.3fms] [rows:%v] %s",
				slowLog,
				float64(elapsed.Nanoseconds())/1e6, "-",
				sql,
			)
		} else {
			self.logger.Warnf(
				"%s\n[%.3fms] [rows:%v] %s",
				slowLog,
				float64(elapsed.Nanoseconds())/1e6,
				rows,
				sql,
			)
		}

	case self.logLevel == gormlog.Info:
		sql, rows := fc()

		if rows == -1 {
			self.logger.Infof(
				"%s\n[%.3fms] [rows:%v] %s",
				float64(elapsed.Nanoseconds())/1e6,
				"-",
				sql,
			)
		} else {
			self.logger.Infof(
				"%s\n[%.3fms] [rows:%v] %s",
				float64(elapsed.Nanoseconds())/1e6,
				rows,
				sql,
			)
		}
	}
}
