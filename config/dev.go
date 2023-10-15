//go:build !k8s

package config

import "os"

var Config = config{
	DB: DBConfig{
		DSN: os.Getenv("MYSQL_DSN"),
	},
	Redis: RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       1,
	},
}
