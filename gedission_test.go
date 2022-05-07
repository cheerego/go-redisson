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
	t.Log(lock.TryLock(-1, 40000))
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

	t.Log(lock1.TryLock(20000, 40000))
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

	t.Log(lock1.TryLock(20000, 40000))
	time.Sleep(10 * time.Second)
	lock1.UnLock()

}
func TestNewGedissonWatchdog(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	gedisson := NewGodisson(rdb)
	lock1 := gedisson.NewRLock("hkn")
	t.Log(lock1.TryLock(40000, -1))
	time.Sleep(1 * time.Minute)
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
