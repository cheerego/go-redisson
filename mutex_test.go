package godisson

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
	"sync/atomic"
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

func singleLockUnlockTest(times int32, variableName string, g *Godisson) error {
	mutex := g.NewMutex("plus_" + variableName)
	a := 0
	wg := sync.WaitGroup{}
	total := int32(0)
	for i := int32(0); i < times; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := mutex.TryLock(1000, -1)
			if err != nil {
				return
			}
			a++
			_, err = mutex.Unlock()
			if err != nil {
				panic("unlock failed")
			}
			atomic.AddInt32(&total, 1)
		}()
	}
	wg.Wait()
	log.Println(variableName, "=", a)
	if int32(a) != total {
		return fmt.Errorf("mutex lock and unlock test failed, %s shoule equal %d,but equal %d", variableName, total, a)
	}
	return nil
}

func TestMutex_LockUnlock(t *testing.T) {
	testCase := []int32{1, 10, 100, 1000, 10000}
	for _, v := range testCase {
		if err := singleLockUnlockTest(v, "variable_1", getGodisson()); err != nil {
			log.Fatalf("err=%v", err)
		}
	}
}

func TestMultiMutex(t *testing.T) {
	testCases := []int32{1, 10, 100, 1000}
	id := 0
	getId := func() int {
		id++
		return id
	}
	for _, v := range testCases {
		wg := sync.WaitGroup{}
		numOfFailures := int32(0)
		for i := int32(0); i < v; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := singleLockUnlockTest(10, fmt.Sprintf("variable_%d", getId()), getGodisson())
				if err != nil {
					t.Logf("test failed,err=%v", err)
					atomic.AddInt32(&numOfFailures, 1)
					return
				}
			}()
			wg.Wait()
		}
		if numOfFailures != 0 {
			t.Fatalf("multi mutex test failed, numOfFailures should equal 0,but equal %d", numOfFailures)
		}
	}
}
