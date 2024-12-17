# adaptivelock

**Install**:
````
go get github.com/intangere/adapativelock
````

**Usage**:
````
sl := adaptivelock.New()
sl.Lock()
// do your operations here
counter++
val, ok := some_map[key]
// ...
sl.Unlock()
````

**What is an adaptive lock and how does it work?**    
In simple terms, this is a lock that races against itself and goes to sleep if it can't lock. The locking mechanism 
uses a CAS operation protected by a spinlock with a very short critical section that yields to another goroutine if it can't lock. 
The sleep is handled by waiting on a channel with a single item buffer of a pre-defined pointer value. 
When Unlock() is called, the channel will atomically signal that the lock is now available, and whichever goroutines happens to wake up 
and read from the channel will have a chance to race and acquire the lock. The average cost of this during high contention or heavier 
critical sections seems to perform better than both a mutex and a spinlock.    

**Why does it work?**    
My guess is that since we effectively put the goroutine to sleep by waiting on the channel, it allows for the go scheduler 
to context switch to threads that actually have work to do rather than threads that have to check if there is work to do. 
That let's us avoid redundant `runtime.Gosched()` calls which means less wasted time on context switching. This in turn 
also leads to vastly less cpu usage than spinlocks and a decrease in CPU usage even compared to a mutex since the busy loop is avoided.     

On average, the overhead of the adaptive lock is higher than a spin lock with less resource usage and is less performant, 
and lower overhead and more performant than a mutex. Long term the adaptive lock on average seems to be the most performant while 
maintaining the least amount of resource usage.    

**How does the performance compare to mutex and spinlocks?**   

The below are benchmarks of `.Lock()` and `.Unlock()` called using `SetParallelism(2000)`.
````
BenchmarkAdaptiveLock-16                100000000              30.85 ns/op
BenchmarkDefaultSpinLock-16             100000000               8.760 ns/op
BenchmarkMutex-16                       100000000              75.93 ns/op
````

The actual overhead of using the adaptive lock for vert short critical sections is higher than a spinlock, but less than half of that of a mutex.    
Once you start adding more contention or hold a lock for more than a few microseconds or the critical section is more intense, the adaptive lock starts to perform better as shown below.    

The following benchmarks are measuring the time of the completing the benchmark for 100,000,000 iterations.   
Lower meaning it took less time overall. This is not measuring the isolated performance of the lock itself nor is that the goal.    
Those numbers are shown above.    

Incrementing a counter and adding to a map ran with 2000 goroutines. 
````
BenchmarkAdaptiveLockLoop-16       	100000000	       5.966 ns/op
BenchmarkMutexLoop-16              	100000000	      13.97 ns/op
BenchmarkSpinLock-16               	100000000	       6.287 ns/op
BenchmarkAdaptiveLockMapLoop-16    	100000000	       7.468 ns/op
BenchmarkMutexMapLoop-16           	100000000	      21.39 ns/op
BenchmarkSpinLockMapLoop-16        	100000000	      11.67 ns/op
````

Same benchmarks with 10000 goroutines
````
BenchmarkAdaptiveLockLoop-16       	100000000	     146.4 ns/op
BenchmarkMutexLoop-16              	100000000	     326.4 ns/op
BenchmarkSpinLock-16               	100000000	     158.9 ns/op
````

The same general pattern can be observed for 1000 goroutines as well.

In terms of CPU usage with my 16 cores, at 10000 goroutines:    
- the adaptive lock had 1 core at 100% usage, with 3-4 at 20-35% usage, the rest 0%.
- the mutex had 1 core at 200%, with all remaining 15 cores at 15-25%. My fans started spinning here.
- the spinlock had 1 core at 2000% with the remaining 15 cores at 100%. I could have fried an egg on my CPU in seconds.    


Note: Initializing a lock using `var al adaptivelock.AdaptiveLock` is not valid because the internal channel is not initialized.
