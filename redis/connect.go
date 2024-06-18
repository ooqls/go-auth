package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/braumsmilk/go-registry"
)

func NewRedis() *redis.Ring {
	redisSrv := registry.Get().Redis
	return redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": redisSrv.GetConnectionString(),
		},
	})
}
