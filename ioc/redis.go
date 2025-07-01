package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
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

func InitRLockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
