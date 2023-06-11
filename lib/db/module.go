package db

import (
	"errors"
	"time"

	"github.com/hung0913208/go-algorithm/lib/container"

	"gorm.io/gorm"
)

type dbModuleImpl struct {
	dbObj   Db
	timeout time.Duration

	initCallback func(module *dbModuleImpl, timeout time.Duration) error
}

var (
	_ container.Module = &dbModuleImpl{}
)

func (self *dbModuleImpl) Init(timeout time.Duration) error {
	self.timeout = timeout
	return self.initCallback(self, timeout*time.Millisecond)
}

func (self *dbModuleImpl) Deinit() error {
	return nil
}

func (self *dbModuleImpl) GetTimeout() time.Duration {
	return self.timeout
}

func NewBigQueryModule(
	projectId, dataset string,
) (container.Module, error) {
	return &dbModuleImpl{
		initCallback: func(module *dbModuleImpl, timeout time.Duration) error {
			dbObj, err := NewBigQuery(projectId, dataset, timeout)
			if err != nil {
				return err
			}

			module.dbObj = dbObj
			return nil
		},
	}, nil
}
func NewMysqlModule(
	host string,
	port int,
	username, password, database string,
) (container.Module, error) {
	return &dbModuleImpl{
		initCallback: func(module *dbModuleImpl, timeout time.Duration) error {
			dbObj, err := NewMysqlV1(host, port, username, password, database,
				timeout, true)
			if err != nil {
				return err
			}

			module.dbObj = dbObj
			return nil
		},
	}, nil
}

func NewPgModule(
	host string,
	port int,
	username, password, database string,
) (container.Module, error) {
	return &dbModuleImpl{
		initCallback: func(module *dbModuleImpl, timeout time.Duration) error {
			dbObj, err := NewPgV1(host, port, username, password, database,
				timeout, "disable",
				true)
			if err != nil {
				return err
			}

			module.dbObj = dbObj
			return nil
		},
	}, nil
}

func NewPgPoolModule(
	host string,
	port int,
	username, password, database string,
) (container.Module, error) {
	return &dbModuleImpl{
		initCallback: func(module *dbModuleImpl, timeout time.Duration) error {
			dbObj, err := NewPgV1(
				host, port, username, password, database,
				timeout, "disable",
				false)
			if err != nil {
				return err
			}

			module.dbObj = dbObj
			return nil
		},
	}, nil
}

func NewPgModuleWithSsl(
	host string,
	port int,
	username, password, database string,
) (container.Module, error) {
	return &dbModuleImpl{
		initCallback: func(module *dbModuleImpl, timeout time.Duration) error {
			dbObj, err := NewPgV1(host, port, username, password, database,
				timeout, "require",
				true)
			if err != nil {
				return err
			}

			module.dbObj = dbObj
			return nil
		},
	}, nil
}

func Establish(module container.Module) (*gorm.DB, error) {
	if wrapper, ok := module.(*dbModuleImpl); ok {
		if wrapper.dbObj != nil {
			return wrapper.dbObj.EstablishV1(), nil
		} else {
			return nil, errors.New("can't establish database connection")
		}
	}
	return nil, errors.New("Can't cast module to database module")
}
