package adaptivelock

import (
	"sync"
	"testing"
	"fmt"
	//"time"
)

var m = map[string]int{
	"1:": 1,
	"2:": 2,
	"3:": 3,
	"4:": 4,
	"5": 5,
	"6:": 6,
	"7": 7,
	"8": 8,
	"9": 9,
	"10": 10,
}

var expected_counter_result int
func init() {
        for j := 0; j < goroutines * goroutines; j++ {
                for _, v := range m {
                        expected_counter_result += v
                }
        }
}

var goroutines = 2000

// weird counter loop test
func LockCounterLoop(sl sync.Locker) {

	var wg sync.WaitGroup
	var counter int

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < goroutines; j++ {
				sl.Lock()
				for _, v := range m {
					counter += v
				}
				sl.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	// verify the counter result
	if counter != expected_counter_result {
		panic(fmt.Sprintf("Result %d; want %d", counter, expected_counter_result))
	}
}

// simple map assignment test which is a slightly better critical section
// also maps are more likely to race so good test in general (ofc there is -race)
func LockMapLoop(sl sync.Locker) {
	mp := map[int]bool{}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < goroutines; j++ {
				sl.Lock()
				for _, v := range m {
					mp[v] = true
				}
				sl.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkAdaptiveLockLoop(t *testing.B) {
	sl := New()
	LockCounterLoop(sl)
}

func BenchmarkMutexLoop(t *testing.B) {
	var sl sync.Mutex
	LockCounterLoop(&sl)
}

func BenchmarkSpinLock(t *testing.B) {
	var sl SpinLock
	LockCounterLoop(&sl)
}

func BenchmarkAdaptiveLockMapLoop(t *testing.B) {
	sl := New()
	LockMapLoop(sl)
}
func BenchmarkMutexMapLoop(t *testing.B) {
	var sl sync.Mutex
	LockMapLoop(&sl)
}
func BenchmarkSpinLockMapLoop(t *testing.B) {
	var sl SpinLock
	LockMapLoop(&sl)
}

func BenchmarkAdaptiveLock(b *testing.B) {
	sl := New()
	b.SetParallelism(goroutines)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sl.Lock()
			_ = 1
			sl.Unlock()
		}
	})
}

func BenchmarkDefaultSpinLock(b *testing.B) {
	var sl SpinLock
	b.SetParallelism(goroutines)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sl.Lock()
			_ = 1
			sl.Unlock()
		}
	})
}

func BenchmarkMutex(b *testing.B) {
	var mu sync.Mutex
	b.SetParallelism(goroutines)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			_ = 1
			mu.Unlock()
		}
	})
}
