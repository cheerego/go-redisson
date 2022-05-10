package examples

import (
	"github.com/cheerego/godisson"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"log"
	"time"
)

func main() {

	// create redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	g := godisson.NewGodisson(rdb, godisson.WithWatchDogTimeout(30*time.Second))

	test1(g)
	test2(g)
}

// can't obtain lock in a same goroutine
func test1(g *godisson.Godisson) {
	m1 := g.NewMutex("godisson")
	m2 := g.NewMutex("godisson")

	err := m1.TryLock(-1, 20000)
	if errors.Is(err, godisson.ErrLockNotObtained) {
		log.Println("can't obtained lock")
	} else if err != nil {
		log.Fatalln(err)
	}
	defer m1.Unlock()

	// because waitTime = -1, waitTime < 0, try once, will return ErrLockNotObtained
	err = m2.TryLock(-1, 20000)
	if errors.Is(err, godisson.ErrLockNotObtained) {
		log.Println("m2 must not obtained lock")
	} else if err != nil {
		log.Fatalln(err)
	}
	time.Sleep(10 * time.Second)
}

func test2(g *godisson.Godisson) {
	m1 := g.NewMutex("godisson")
	m2 := g.NewMutex("godisson")

	go func() {
		err := m1.TryLock(-1, 20000)
		if errors.Is(err, godisson.ErrLockNotObtained) {
			log.Println("can't obtained lock")
		} else if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(10 * time.Second)
		m1.Unlock()
	}()

	// waitTime > 0, after 10 milliseconds will obtain the lock
	go func() {
		time.Sleep(1 * time.Second)

		err := m2.TryLock(15000, 20000)
		if errors.Is(err, godisson.ErrLockNotObtained) {
			log.Println("m2 must not obtained lock")
		} else if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(10 * time.Second)

		m2.Unlock()
	}()
	time.Sleep(20 * time.Second)

}
