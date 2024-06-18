package usercachev1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/braumsmilk/go-log"
	"github.com/braumsmilk/go-auth/redis"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("userNameCache")

type UserNameCache interface {
	GetUser(id int) string
	AddUser(id int, name string)
}

func NewRedisCache() *RedisCache {
	r := redis.NewRedis()
	return &RedisCache{
		c: *cache.New(&cache.Options{
			Redis:      r,
			LocalCache: cache.NewTinyLFU(100, time.Hour),
		}),
	}
}

type RedisCache struct {
	c cache.Cache
}

func (r *RedisCache) GetUser(id int) string {
	key := fmt.Sprintf("username-%d", id)
	var name string
	err := r.c.Get(context.Background(), key, &name)
	if err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			return ""
		}
		l.Warn("failed to get username from redis cache", zap.Error(err))
		return ""
	}

	return name
}

func (r *RedisCache) AddUser(id int, name string) {
	err := r.c.Set(&cache.Item{
		Key:   fmt.Sprintf("username-%d", id),
		Value: name,
	})
	if err != nil {
		l.Warn("failed to update redis cache with username", zap.Error(err))
	}
}
