package godisson

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"time"
)

var ErrLockNotObtained = errors.New("ErrLockNotObtained")

const UNLOCK_MESSAGE int64 = 0

const READ_UNLOCK_MESSAGE int64 = 1

type RLock struct {
	Key string
	g   *Godisson
}

func newRLock(key string, g *Godisson) *RLock {
	return &RLock{Key: key, g: g}
}

func (r *RLock) Success() {

}
func (r *RLock) Lock() error {
	return r.TryLock(-1, -1)
}

//TryLock try to obtain lock
// waitTime, Millisecond
// leaseTime, Millisecond, -1 enable watchdog
func (r *RLock) TryLock(waitTime int64, leaseTime int64) error {
	wait := waitTime
	current := currentTimeMillis()
	ttl, err := r.tryAcquire(waitTime, leaseTime)
	if err != nil {
		return err
	}
	if ttl == 0 {
		return nil
	}
	wait -= currentTimeMillis() - current
	if wait <= 0 {
		return ErrLockNotObtained
	}
	current = currentTimeMillis()
	// PubSub
	sub := r.g.c.Subscribe(context.TODO(), r.getChannelName())
	defer sub.Close()
	timeoutCtx, timeoutCancel := context.WithTimeout(context.TODO(), time.Duration(wait)*time.Millisecond)
	defer timeoutCancel()
	_, err = sub.ReceiveMessage(timeoutCtx)
	if err != nil {
		return ErrLockNotObtained
	}

	wait -= currentTimeMillis() - current
	if wait <= 0 {
		return ErrLockNotObtained
	}

	for {
		currentTime := currentTimeMillis()
		ttl, err = r.tryAcquire(waitTime, leaseTime)
		if ttl == 0 {
			return nil
		}
		wait -= currentTimeMillis() - currentTime
		if wait <= 0 {
			return ErrLockNotObtained
		}
		currentTime = currentTimeMillis()

		var target *net.OpError
		if ttl >= 0 && ttl < wait {
			tCtx, _ := context.WithTimeout(context.TODO(), time.Duration(ttl)*time.Millisecond)
			_, err := sub.ReceiveMessage(tCtx)
			if err != nil {
				if errors.As(err, &target) {
					continue
				}
			}
		} else {
			tCtx, _ := context.WithTimeout(context.TODO(), time.Duration(wait)*time.Millisecond)
			_, err := sub.ReceiveMessage(tCtx)

			if err != nil {
				if errors.As(err, &target) {
					continue
				}
			}
		}
		wait -= currentTimeMillis() - currentTime
		if wait <= 0 {
			return ErrLockNotObtained
		}
	}
}

func (r *RLock) tryAcquire(waitTime int64, leaseTime int64) (int64, error) {
	goid, err := gid()
	if err != nil {
		return 0, err
	}
	if leaseTime != -1 {
		return r.tryAcquireInner(waitTime, leaseTime)
	}
	// watch dog
	ttl, err := r.tryAcquireInner(waitTime, r.g.watchDogTimeout.Milliseconds())
	if err != nil {
		return 0, nil
	}
	if ttl == 0 {
		r.renewScheduler(goid)
	}
	return ttl, err
}

func (r *RLock) renewScheduler(goroutineId uint64) {
	newEntry := NewRenewEntry()
	entryName := r.g.getEntryName(r.Key)
	if oldEntry, ok := r.g.RenewMap.Get(entryName); ok {
		oldEntry.(*RenewEntry).addGoroutineId(goroutineId)
	} else {
		newEntry.addGoroutineId(goroutineId)
		cancel, cancelFunc := context.WithCancel(context.TODO())

		go func(context.Context) {
			ticker := time.NewTicker(r.g.watchDogTimeout / 3)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					renew, err := r.renew(goroutineId)
					if err != nil {
						return
					}
					// key not exists, so return goroutine
					if renew == 0 {
						return
					}
				case <-cancel.Done():
					return
				}
			}
		}(cancel)

		newEntry.cancelFunc = cancelFunc
		r.g.RenewMap.Set(entryName, newEntry)
	}

}

func (r *RLock) cancelExpirationRenewal(goid uint64) {
	//entryName := r.g.getEntryName(r.Key)
	//if entry, ok := r.g.RenewMap.Get(entryName); ok {
	//	r.g.RenewMap.Remove(entryName)
	//}
}

func (r *RLock) tryAcquireInner(waitTime int64, leaseTime int64) (int64, error) {
	gid, err := gid()
	if err != nil {
		return 0, err
	}
	lockName := r.g.getLockName(gid)
	if err != nil {
		return 0, err
	}
	result, err := r.g.c.Eval(context.TODO(), `
if (redis.call('exists', KEYS[1]) == 0) then
    redis.call('hincrby', KEYS[1], ARGV[2], 1);
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return 0;
end;
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
    redis.call('hincrby', KEYS[1], ARGV[2], 1);
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return 0;
end;
return redis.call('pttl', KEYS[1]);
`, []string{r.Key}, leaseTime, lockName).Result()
	if err != nil {
		return 0, err
	}

	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, errors.Errorf("try lock result converter to int64 error, value is %v", result)
	}

}

func (r *RLock) UnLock() (int64, error) {
	goid, err := gid()
	if err != nil {
		return 0, err
	}

	defer func() {
		r.cancelExpirationRenewal(goid)
	}()

	result, err := r.g.c.Eval(context.Background(), `
if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then
    return nil;
end;
local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1);
if (counter > 0) then
    redis.call('pexpire', KEYS[1], ARGV[2]);
    return 0;
else
    redis.call('del', KEYS[1]);
    redis.call('publish', KEYS[2], ARGV[1]);
    return 1;
end;
return nil;
`, []string{r.Key, r.getChannelName()}, UNLOCK_MESSAGE, DefaultWatchDogTimeout, r.g.getLockName(goid)).Result()
	if err != nil {
		return 0, err
	}
	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, errors.Errorf("try lock result converter to bool error, value is %v", result)
	}
}

func (r *RLock) getChannelName() string {
	return fmt.Sprintf("{gedisson_lock__channel}:%s)", r.Key)
}

func (r *RLock) renew(gid uint64) (int64, error) {
	result, err := r.g.c.Eval(context.TODO(), `
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return 1;
end ;
return 0
`, []string{r.Key}, r.g.watchDogTimeout.Milliseconds(), r.g.getLockName(gid)).Result()
	if err != nil {
		return 0, err
	}
	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, errors.Errorf("try lock result converter to int64 error, value is %v", result)
	}
}
