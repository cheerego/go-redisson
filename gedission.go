package godisson

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	cmap "github.com/orcaman/concurrent-map"
	uuid "github.com/satori/go.uuid"
	"log"
	"time"
)

type Godisson struct {
	c               *redis.Client
	watchDogTimeout time.Duration
	uuid            string
	// key uuid:key, value entry
	RenewMap cmap.ConcurrentMap
}

var DefaultWatchDogTimeout = 30 * time.Second

func NewGodisson(redisClient *redis.Client, opts ...OptionFunc) *Godisson {
	g := &Godisson{
		c:               redisClient,
		uuid:            uuid.NewV4().String(),
		watchDogTimeout: DefaultWatchDogTimeout,
		RenewMap:        cmap.New(),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

type OptionFunc func(g *Godisson)

func WithWatchDogTimeout(t time.Duration) OptionFunc {
	return func(g *Godisson) {
		if t.Seconds() < 30 {
			t = DefaultWatchDogTimeout
			log.Println("watchDogTimeout is too small, so config default ")
		}
		g.watchDogTimeout = t
	}
}

func (g *Godisson) NewRLock(key string) *RLock {
	return newRLock(key, g)
}

func (g *Godisson) NewMutex(key string) *Mutex {
	return newMutex(key, g)
}

func (g *Godisson) getEntryName(key string) string {
	return fmt.Sprintf("%s:%s", g.uuid, key)
}

func (g *Godisson) getChannelName(key string) string {
	return fmt.Sprintf("{gedisson_lock__channel}:%s)", key)
}
