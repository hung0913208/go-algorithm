package background

import (
	"time"

	"github.com/hung0913208/go-algorithm/lib/container"
)

type bgModuleImpl struct {
	timeout time.Duration
}

type Handler func() error

var (
	_ container.Module = &bgModuleImpl{}
)

func (self *bgModuleImpl) Init(timeout time.Duration) error {
	self.timeout = timeout

	return nil
}

func (self *bgModuleImpl) Deinit() error {
	return nil
}

func (self *bgModuleImpl) GetTimeout() time.Duration {
	return self.timeout
}
