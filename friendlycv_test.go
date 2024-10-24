package friendlycv

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const waitersCount = 10
const testDelay = 2 * time.Second

func TestFriendlyCV_Signal(t *testing.T) {
	t.Run("it wakes up exactly one waiter", func(t *testing.T) {
		cv := Cond{L: &sync.Mutex{}}
		ctx, cancel := context.WithTimeout(context.Background(), testDelay)
		defer cancel()

		var wakedUpCounter atomic.Uint32
		var wg sync.WaitGroup
		var wgRun sync.WaitGroup
		wgRun.Add(waitersCount)
		for range waitersCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cv.L.Lock()
				wgRun.Done()
				ok := cv.Wait(ctx)
				if ok {
					wakedUpCounter.Add(1)
					cv.L.Unlock()
				}
			}()
		}

		wgRun.Wait() // Ensure that the waiters are waiting.
		cv.Signal()
		wg.Wait() // Ensure all waiters exited.
		if wakedUpCounter.Load() != 1 {
			t.Errorf("Signal should have awake only one waiter, but %d were awakened", wakedUpCounter.Load())
		}
	})

	t.Run("it will wakes up each waiter one by one", func(t *testing.T) {
		cv := Cond{L: &sync.Mutex{}}
		ctx, cancel := context.WithTimeout(context.Background(), testDelay)
		defer cancel()

		var wakedUpCounter atomic.Uint32
		var wg sync.WaitGroup
		var wgRun sync.WaitGroup
		wgRun.Add(waitersCount)
		for range waitersCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cv.L.Lock()
				wgRun.Done()
				ok := cv.Wait(ctx)
				if ok {
					wakedUpCounter.Add(1)
					cv.L.Unlock()
				}
			}()
		}

		wgRun.Wait() // Ensure that the waiters are waiting.
		for range waitersCount {
			cv.Signal()
		}
		wg.Wait() // Ensure all waiters exited.
		if wakedUpCounter.Load() != wakedUpCounter.Load() {
			t.Errorf("Signal should have awake %d waiters, but %d were awakened", waitersCount, wakedUpCounter.Load())
		}
	})
}

func TestFriendlyCV_Broadcast(t *testing.T) {
	t.Run("it wakes up all waiters", func(t *testing.T) {
		cv := Cond{L: &sync.Mutex{}}
		ctx, cancel := context.WithTimeout(context.Background(), testDelay)
		defer cancel()

		var wakedUpCounter atomic.Uint32
		var wg sync.WaitGroup
		var wgRun sync.WaitGroup
		wgRun.Add(waitersCount)
		for range waitersCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cv.L.Lock()
				wgRun.Done()
				ok := cv.Wait(ctx)
				if ok {
					wakedUpCounter.Add(1)
					cv.L.Unlock()
				}
			}()
		}

		wgRun.Wait() // Ensure that the waiters are waiting.
		cv.Broadcast()
		wg.Wait() // Ensure all waiters exited.
		if wakedUpCounter.Load() != waitersCount {
			t.Errorf("Signal should have awake %d waiter, but %d were awakened", waitersCount, wakedUpCounter.Load())
		}
	})
}

func TestFriendlyCV_Wait(t *testing.T) {
	t.Run("it waits until being awakened", func(t *testing.T) {
		cv := Cond{L: &sync.Mutex{}}
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		go func() {
			time.Sleep(time.Second)
			cv.Broadcast()
		}()

		cv.L.Lock()
		result := cv.Wait(ctx)
		if !result {
			t.Error("Wait should have returned true, but returned false")
		}
	})

	t.Run("it waits until the context become cancelled", func(t *testing.T) {
		cv := Cond{L: &sync.Mutex{}}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		cv.L.Lock()
		result := cv.Wait(ctx)
		if result {
			t.Error("Wait should have returned false, but returned true")
		}
	})
}
