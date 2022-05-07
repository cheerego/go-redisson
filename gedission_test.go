package godisson

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"net"
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
	lock := gedisson.NewRLock("hkn")
	t.Log(lock.TryLock(context.Background(), -1*time.Second, 40*time.Second))
	time.Sleep(10 * time.Second)
	lock.UnLock()

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

	t.Log(lock1.TryLock(context.Background(), 20*time.Second, 40*time.Second))
	time.Sleep(10 * time.Second)
	lock1.UnLock()

}

func TestNewGedisson2(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	gedisson := NewGodisson(rdb)
	lock1 := gedisson.NewRLock("hkn")

	t.Log(lock1.TryLock(context.Background(), 20*time.Second, 40*time.Second))
	time.Sleep(10 * time.Second)
	lock1.UnLock()

}

func TestSub(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	sub := rdb.Subscribe(context.Background(), "123-channel")
	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()

	message, err := sub.ReceiveMessage(ctx)
	var target *net.OpError
	t.Log(message, err, errors.As(err, &target))
}

func TestPub(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	rdb.Publish(context.Background(), "123-channel", "123")

}
