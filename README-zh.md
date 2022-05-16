## go-redisson 

一个借鉴 redisson 用 go 实现的 Redis 分布式锁的库。

## 安装

```shell
go get github.com/cheerego/go-redisson
```


## 支持锁类型

* Mutex [示例](#Mutex) 特性：
  * 互斥锁 (X Lock)。 
  * 和 go 标准库 sync.Mutex 用起来差不多。 
  * 不支持可重入。 
  * 支持 WatchDog。

* RLock[示例](#RLock) 特性：
  * 互斥可重入锁。 
  * 和 redisson 使用起来差不多。 
  * 支持同一个协程重复加锁。
  * 支持 WatchDog。

## 特性

* tryLock，if waitTime > 0, wait `waitTime` milliseconds to try to obtain lock by while true and redis pub sub.
* watchdog, if leaseTime = -1, start a time.Ticker(defaultWatchDogTime / 3) to renew lock expiration time.

## 配置 

### WatchDogTimeout

```go
g := godisson.NewGodisson(rdb, godisson.WithWatchDogTimeout(30*time.Second))
```


## 示例

### Mutex 

```go
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

}

```


### RLock
```go
package examples

import (
	"github.com/cheerego/godisson"
	"github.com/go-redis/redis/v8"
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

	// lock with watchdog without retry
	lock := g.NewRLock("godisson")

	err := lock.Lock()
	if err == godisson.ErrLockNotObtained {
		log.Println("Could not obtain lock")
	} else if err != nil {
		log.Fatalln(err)
	}
	defer lock.Unlock()

	// lock with retry、watchdog
	// leaseTime value is -1, enable watchdog
	lock2 := g.NewRLock("godission-try-watchdog")

	err = lock2.TryLock(20000, -1)
	if err == godisson.ErrLockNotObtained {
		log.Println("Could not obtain lock")
	} else if err != nil {
		log.Fatalln(err)
	}
	time.Sleep(10 * time.Second)
	defer lock.Unlock()
}

```