package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	c              *redis.Client
	expirationTime time.Duration
}

func New(c *redis.Client, expirationTime time.Duration) *redisCache {
	return &redisCache{
		c:              c,
		expirationTime: expirationTime,
	}
}
