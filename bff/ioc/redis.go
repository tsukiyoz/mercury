package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

//go:generate mockgen -package=redismocks -destination=../internal/repository/mocks/cache/redis/cmdable.mock.go github.com/redis/go-redis/v9 Cmdable
func InitRedis() redis.Cmdable {
	type Config struct {
		Addr string `yaml:"addr"`
	}

	var cfg Config
	viper.UnmarshalKey("redis", &cfg)

	cmd := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})
	return cmd
}
