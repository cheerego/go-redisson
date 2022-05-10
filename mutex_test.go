package godisson

import (
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

func getGodisson() *Godisson {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return NewGodisson(rdb)
}

func TestMutex_TryLock(t *testing.T) {
	g := getGodisson()
	mutex := g.NewMutex("hkn")

	t.Log(mutex.TryLock(-1, -1))

	time.Sleep(30 * time.Second)
	mutex.Unlock()

}

func TestMutex_TryLock2(t *testing.T) {
	g := getGodisson()
	mutex := g.NewMutex("hkn")

	t.Log(mutex.TryLock(50000, -1))

	time.Sleep(30 * time.Second)
	mutex.Unlock()

}
