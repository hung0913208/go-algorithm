package kv

import (
	"context"
	"errors"
	"time"

	mdb "github.com/memcachier/mc/v3"
	rdbOtel "github.com/redis/go-redis/extra/redisotel/v9"
	rdb "github.com/redis/go-redis/v9"
)

type KV interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Expire(ctx context.Context, key string, timeout time.Duration) error

	Inc(ctx context.Context, key string) (int64, error)
}

type kvTypeEnum int

const (
	Unknown kvTypeEnum = iota
	Redis
	Memcache
)

type kvImpl struct {
	kvType  kvTypeEnum
	rdbConn *rdb.Client
	mdbConn *mdb.Client
}

func NewRedis(url string) (KV, error) {
	opt, err := rdb.ParseURL(url)
	if err != nil {
		return nil, err
	}

	rdbConn := rdb.NewClient(opt)
	if err := rdbOtel.InstrumentTracing(rdbConn); err != nil {
		return nil, err
	}
	if err := rdbOtel.InstrumentMetrics(rdbConn); err != nil {
		return nil, err
	}

	return &kvImpl{
		rdbConn: rdbConn,
		kvType:  Redis,
	}, nil
}

func NewMemcache(server, username, password string) (KV, error) {
	return &kvImpl{
		mdbConn: mdb.NewMC(server, username, password),
		kvType:  Memcache,
	}, nil
}

func (self *kvImpl) Get(
	ctx context.Context,
	key string,
) (string, error) {
	switch self.kvType {
	case Redis:
		return self.rdbConn.Get(ctx, key).Result()
	case Memcache:
		val, _, _, err := self.mdbConn.Get(key)
		if err != nil {
			return "", err
		}
		return val, err
	default:
		return "", errors.New("Not support this format")
	}
}

func (self *kvImpl) Set(
	ctx context.Context,
	key, value string,
) error {
	switch self.kvType {
	case Redis:
		return self.rdbConn.Set(ctx, key, value, 0).Err()
	case Memcache:
		_, err := self.mdbConn.Set(key, value, 0, 0, 0)
		if err != nil {
			return err
		}

		return nil
	default:
		return errors.New("Not support this format")
	}
}

func (self *kvImpl) Expire(
	ctx context.Context,
	key string,
	timeout time.Duration,
) error {
	switch self.kvType {
	case Redis:
		_, err := self.rdbConn.Expire(ctx, key, timeout).
			Result()
		return err
	case Memcache:
		value, err := self.Get(ctx, key)
		if err != nil {
			return err
		}

		_, err = self.mdbConn.Set(key, value,
			0,
			uint32(timeout/time.Second),
			0)
		if err != nil {
			return err
		}

		return nil
	default:
		return errors.New("Not support this format")
	}
}

func (self *kvImpl) Inc(
	ctx context.Context,
	key string,
) (int64, error) {
	switch self.kvType {
	case Redis:
		return self.rdbConn.Incr(ctx, key).
			Result()
	default:
		return 0, errors.New("Not support this format")
	}
}
