package startup

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type RedisConfig struct {
		addr string `yaml:"addr"`
	}
	var redisCfg = RedisConfig{
		addr: "localhost:6379",
	}
	err := viper.UnmarshalKey("redis", &redisCfg)
	if err != nil {
		panic(err)
	}
	return redis.NewClient(&redis.Options{
		Addr: redisCfg.addr,
	})
}
