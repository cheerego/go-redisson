package godisson

import (
	"context"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

func TestNewGedisson(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	gedisson := NewGodisson(rdb)
	t.Log(gedisson.NewRLock("hkn").TryLock(context.Background(), -1, 30*time.Second))
	t.Log(gedisson.NewRLock("hkn").TryLock(context.Background(), -1, 30*time.Second))

	//time.Sleep(100 * time.Second)
}

func TestNewGedisson1(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	gedisson := NewGodisson(rdb)
	lock1 := gedisson.NewRLock("hkn")
	lock2 := gedisson.NewRLock("hkn")

	t.Log(lock1.TryLock(context.Background(), -1, 30*time.Second))
	t.Log(lock1)
	t.Log(lock2.TryLock(context.Background(), -1, 30*time.Second))
	t.Log(lock2)

}

func TestSub(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	sub := rdb.Subscribe(context.Background(), "123-channel")
	channel := sub.Channel()

	for message := range channel {
		t.Log(message)
	}
}

func TestPub(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	rdb.Publish(context.Background(), "123-channel", "123")

}