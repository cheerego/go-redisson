package godisson

import (
	"context"
	"github.com/pkg/errors"
	"net"
	"time"
)

type Mutex struct {
	Key        string
	g          *Godisson
	cancelFunc context.CancelFunc
}

var _ Locker = (*Mutex)(nil)

func newMutex(key string, g *Godisson) *Mutex {
	return &Mutex{Key: key, g: g}
}

func (m *Mutex) TryLock(waitTime int64, leaseTime int64) error {
	wait := waitTime
	current := currentTimeMillis()
	ttl, err := m.tryAcquire(waitTime, leaseTime)
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
	sub := m.g.c.Subscribe(context.TODO(), m.g.getChannelName(m.Key))
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
		ttl, err = m.tryAcquire(waitTime, leaseTime)
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

func (m *Mutex) tryAcquire(waitTime int64, leaseTime int64) (int64, error) {

	if leaseTime != -1 {
		return m.tryAcquireInner(waitTime, leaseTime)
	}

	ttl, err := m.tryAcquire(waitTime, m.g.watchDogTimeout.Milliseconds())
	if err != nil {
		return 0, err
	}
	if ttl == 0 {
		m.renewExpirationScheduler()
	}
	return ttl, nil
}

func (m *Mutex) tryAcquireInner(waitTime int64, leaseTime int64) (int64, error) {
	result, err := m.g.c.Eval(context.Background(), `
if (redis.call('exists', KEYS[1]) == 0) then
	redis.call('set', KEYS[1], 1);
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return 0;
end;
return redis.call('pttl', KEYS[1]);
`, []string{m.Key}, leaseTime).Result()
	if err != nil {
		return 0, err
	}

	if ttl, ok := result.(int64); ok {
		return ttl, nil
	} else {
		return 0, errors.Errorf("tryAcquireInner result converter to int64 error, value is %v", result)
	}
}

func (m *Mutex) Unlock() (int64, error) {
	defer m.cancelExpirationRenewal()
	result, err := m.g.c.Eval(context.TODO(), `
if (redis.call('exists', KEYS[1]) == 0) then
    return 0;
end;
redis.call('del', KEYS[1]);
redis.call('publish', KEYS[2], ARGV[1]);
return 1;
`, []string{m.Key, m.g.getChannelName(m.Key)}, UNLOCK_MESSAGE).Result()
	if err != nil {
		return 0, err
	}

	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, errors.Errorf("unlock result converter to bool error, value is %v", result)
	}
}

func (m *Mutex) renewExpirationScheduler() {
	cancel, cancelFunc := context.WithCancel(context.TODO())
	m.cancelFunc = cancelFunc
	go m.renewExpirationSchedulerGoroutine(cancel)
}

func (m *Mutex) renewExpirationSchedulerGoroutine(cancel context.Context) {

	ticker := time.NewTicker(m.g.watchDogTimeout / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			renew, err := m.renewExpiration()
			if err != nil {
				return
			}
			// key not exists, so return goroutine
			if renew == 0 {
				m.cancelExpirationRenewal()
				return
			}
		case <-cancel.Done():
			return
		}
	}
}

func (m *Mutex) renewExpiration() (int64, error) {
	result, err := m.g.c.Eval(context.TODO(), `
if (redis.call('exists', KEYS[1]) == 1) then
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return 1;
end;
return 0;
`, []string{m.Key}, m.g.watchDogTimeout.Milliseconds()).Result()
	if err != nil {
		return 0, err
	}
	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, errors.Errorf("try lock result converter to int64 error, value is %v", result)
	}
}

func (m *Mutex) cancelExpirationRenewal() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}
