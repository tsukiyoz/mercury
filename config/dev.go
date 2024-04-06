//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:for.nothing@tcp(127.0.0.1:3306)/mercury",
	},
	Redis: RedisConfig{
		Addr:     "localhost:6379",
		Password: "for.nothing",
		DB:       1,
	},
}
