package kv

import (
	"errors"
	"time"

	"github.com/hung0913208/go-algorithm/lib/container"
)

type kvModuleImpl struct {
	kvObj   KV
	timeout time.Duration

	initCallback func(module *kvModuleImpl, timeout time.Duration) error
}

var (
	_ container.Module = &kvModuleImpl{}
)

func (self *kvModuleImpl) Init(timeout time.Duration) error {
	self.timeout = timeout
	return self.initCallback(self, timeout*time.Millisecond)
}

func (self *kvModuleImpl) Deinit() error {
	return nil
}

func (self *kvModuleImpl) GetTimeout() time.Duration {
	return self.timeout
}

func NewRedisModule(uri string) (container.Module, error) {
	return &kvModuleImpl{
		initCallback: func(module *kvModuleImpl, timeout time.Duration) error {
			kvObj, err := NewRedis(uri)
			if err != nil {
				return err
			}

			module.kvObj = kvObj
			return nil
		},
	}, nil
}

func NewMemcacheModule(
	server, username, password string,
) (container.Module, error) {
	return &kvModuleImpl{
		initCallback: func(module *kvModuleImpl, timeout time.Duration) error {
			kvObj, err := NewMemcache(server, username, password)
			if err != nil {
				return err
			}

			module.kvObj = kvObj
			return nil
		},
	}, nil
}

func Establish(module container.Module) (KV, error) {
	if wrapper, ok := module.(*kvModuleImpl); ok {
		if wrapper.kvObj != nil {
			return wrapper.kvObj, nil
		} else {
			return nil, errors.New("can't establish database connection")
		}
	}
	return nil, errors.New("Can't cast module to database module")
}
