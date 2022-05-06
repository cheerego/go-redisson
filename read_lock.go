package godisson

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

var ErrNotObtain = errors.New("ErrNotObtain")

const UNLOCK_MESSAGE int64 = 0

const READ_UNLOCK_MESSAGE int64 = 1

type RLock struct {
	Key      string
	c        *redis.Client
	lockName string
	uuid     string
}

func newRLock(key string, uuid string, c *redis.Client) *RLock {
	return &RLock{Key: key, uuid: uuid, c: c}
}

func (r *RLock) Success() {

}
func (r *RLock) Lock(ctx context.Context) error {
	return r.TryLock(ctx, -1, -1)
}

func (r *RLock) TryLock(ctx context.Context, waitTime time.Duration, leaseTime time.Duration) error {
	time := waitTime.Milliseconds()
	current := CurrentTimeMillis()
	ttl, err := r.tryAcquire(ctx, waitTime, leaseTime)
	if err != nil {
		return err
	}
	if ttl == 0 {
		return nil
	}
	time -= CurrentTimeMillis() - current
	if time <= 0 {
		return ErrNotObtain
	}
	current = CurrentTimeMillis()
	// PubSub

	time -= -current
	if time <= 0 {
		return ErrNotObtain
	}
	return nil
}

func (r *RLock) tryAcquire(ctx context.Context, waitTime time.Duration, leaseTime time.Duration) (int64, error) {
	lockName, err := GetLockName(r.uuid)
	if err != nil {
		return 0, err
	}
	r.lockName = lockName
	result, err := r.c.Eval(ctx, `
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
`, []string{r.Key}, leaseTime.Milliseconds(), lockName).Result()
	if err != nil {
		return 0, err
	}

	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, errors.Errorf("try lock result converter to bool error, value is %v", result)
	}
}

func (r *RLock) UnLock() (bool, error) {
	lockName, err := GetLockName(r.uuid)
	if err != nil {
		return false, err
	}
	result, err := r.c.Eval(context.Background(), `
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
`, []string{
		r.Key, r.getChannelName()}, UNLOCK_MESSAGE, DefaultLockTTLTime, lockName).Result()
	if err != nil {
		return false, err
	}
	if b, ok := result.(bool); ok {
		return b, nil
	} else {
		return false, errors.Errorf("try lock result converter to bool error, value is %v", result)
	}
}

func (r *RLock) getChannelName() string {
	return fmt.Sprintf("{gedisson_lock__channel}:%s)", r.Key)
}
