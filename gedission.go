package godisson

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"time"
)

type Godisson struct {
	c    *redis.Client
	uuid string
}

var DefaultLockTTLTime = 30 * time.Second

func NewGodisson(redisClient *redis.Client) *Godisson {
	return &Godisson{c: redisClient, uuid: uuid.New().String()}
}

func (g *Godisson) NewRLock(key string) *RLock {
	return newRLock(key, g.uuid, g.c)
}
