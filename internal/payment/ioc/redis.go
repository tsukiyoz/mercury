package ioc

import (
	"context"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

//go:generate mockgen -package=redismocks -destination=../repository/mocks/cache/redis/cmdable.mock.go github.com/redis/go-redis/v9 Cmdable
func InitRedis() redis.Cmdable {
	type Config struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	}

	var cfg Config
	viper.UnmarshalKey("redis", &cfg)

	cmd := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	pingCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp := cmd.Ping(pingCtx)
	if resp.Err() != nil {
		panic(resp.Err())
	}
	return cmd
}

func InitRLockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
