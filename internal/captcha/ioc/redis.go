package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

//go:generate mockgen -package=redismocks -destination=../repository/cache/mocks/redis/cmdable.mock.go github.com/redis/go-redis/v9 Cmdable
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
	return cmd
}
