package logs

import (
	"fmt"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
)

type errorImpl struct {
	logger  Logger
	timeout time.Duration
}

func NewErrorModule() container.Module {
	return &errorImpl{
		logger: NewLoggerWithStacktrace(),
	}
}

func (self *errorImpl) Init(timeout time.Duration) error {
	self.timeout = timeout
	return nil
}

func (self *errorImpl) Deinit() error {
	return nil
}

func (self *errorImpl) GetTimeout() time.Duration {
	return self.timeout
}

func Errorf(format string, args ...interface{}) error {
	for _, arg := range args {
		if arg == nil {
			return nil
		}
	}

	errorModule, err := container.Pick("error")
	if err != nil {
		return err
	}

	if wrapper, ok := errorModule.(*errorImpl); !ok {
		return fmt.Errorf("not found module `error`")
	} else {
		wrapper.logger.Errorf(format, args...)
		return fmt.Errorf(format, args...)
	}
}

func Error(err error) error {
	if err == nil {
		return nil
	}

	return Errorf("%v", err)
}
