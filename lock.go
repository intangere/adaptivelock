package adaptivelock

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	locked atomic.Bool
}

func (sl *SpinLock) TryLock() bool {
	return sl.locked.CompareAndSwap(false, true)
}

func (sl *SpinLock) Lock() {
	for !sl.TryLock() {
		runtime.Gosched()
	}
}

func (sl *SpinLock) Unlock() {
	sl.locked.Store(false)
}

type AdaptiveLock struct {
	ch      chan *struct{}
	sl      SpinLock
	rw      SpinLock
	holders atomic.Int32
	sent    atomic.Bool
}

func New() *AdaptiveLock {
	return &AdaptiveLock{
		ch: make(chan *struct{}, 1),
	}
}

var wakeup = &struct{}{}

func (al *AdaptiveLock) Lock() {

	for {
		al.rw.Lock()
		if al.sl.TryLock() {
			al.rw.Unlock()
			return
		}

		al.holders.Add(1)
		al.rw.Unlock()

		<-al.ch

		al.rw.Lock()
		al.sent.Store(false)
		al.holders.Add(-1)
		al.rw.Unlock()
	}
}

func (al *AdaptiveLock) Unlock() {
	al.rw.Lock()
	if al.holders.Load() > 0 && !al.sent.Load() {
		al.holders.Add(1)
		al.sent.Store(true)
		al.ch <- wakeup // faster to send here
	}
	al.sl.Unlock()
	al.rw.Unlock()
}
