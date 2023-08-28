//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:for.nothing@tcp(124.70.190.134:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "124.70.190.134:6379",
	},
}
