package godisson

import (
	"context"
	"sync"
)

type RenewEntry struct {
	sync.Mutex
	goroutineIds map[uint64]int64
	cancelFunc   context.CancelFunc
}

func NewRenewEntry() *RenewEntry {
	return &RenewEntry{
		goroutineIds: make(map[uint64]int64),
	}
}

func (r *RenewEntry) addGoroutineId(goroutineId uint64) {
	r.Lock()
	defer r.Unlock()
	count, ok := r.goroutineIds[goroutineId]
	if ok {
		count++
	} else {
		count = 1
	}
	r.goroutineIds[goroutineId] = count
}

func (r *RenewEntry) removeGoroutineId(goroutineId uint64) {
	r.Lock()
	defer r.Unlock()

	count, ok := r.goroutineIds[goroutineId]
	if !ok {
		return
	}
	count--
	if count == 0 {
		delete(r.goroutineIds, goroutineId)
	} else {
		r.goroutineIds[goroutineId] = count
	}
}

func (r *RenewEntry) hasNoThreads() bool {
	return len(r.goroutineIds) == 0
}
