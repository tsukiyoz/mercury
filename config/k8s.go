//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:for.nothing@tcp(mysql-service:3308)/webook",
	},
	Redis: RedisConfig{
		Addr:     "redis-service:6380",
		Password: "",
		DB:       1,
	},
}
