package godisson

type Locker interface {
	TryLock(waitTime int64, leaseTime int64) error
	Unlock() (int64, error)
}
