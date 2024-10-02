//go:build !k8s

package config

var Config = config{
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
	DB: DBConfig{
		// 本地连接
		DSN: "localhost:13316",
	},
}
