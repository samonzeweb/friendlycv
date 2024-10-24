package friendlycv

import (
	"context"
	"sync"
)

// Cond implements a condition variable similar to the one in
// the standard sync package, but managing context cancellations
// while waiting in the [Cond.Wait] method.
// The implementation is certainly less efficient than the sync
// one.
type Cond struct {
	L            sync.Locker
	internalLock sync.Mutex
	waitChans    []chan struct{}
}

// NewCond returns a new Cond with Locker l.
func NewCond(l sync.Locker) *Cond {
	return &Cond{L: l}
}

// Signal wakes up one goroutine waiting on c, it there is any.
func (c *Cond) Signal() {
	c.internalLock.Lock()
	defer c.internalLock.Unlock()

	if len(c.waitChans) == 0 {
		return
	}

	first := c.waitChans[0]
	c.waitChans = c.waitChans[1:]
	close(first)
}

// Broadcast wakes up all goroutines waiting on c.
func (c *Cond) Broadcast() {
	c.internalLock.Lock()
	defer c.internalLock.Unlock()

	for _, ch := range c.waitChans {
		close(ch)
	}
	c.waitChans = c.waitChans[:0]
}

// Wait acts like the Wait method of sync.Cond (see its documentation),
// with two major differences.
// It accepts a context and manage its possible cancellation.
// It returns true if it was awakened by [Cond.Signal] or [Cond.Broadcast],
// otherwise false.
func (c *Cond) Wait(ctx context.Context) bool {
	newWaitChan := make(chan struct{})
	c.internalLock.Lock()
	c.waitChans = append(c.waitChans, newWaitChan)
	c.internalLock.Unlock()

	c.L.Unlock()
	select {
	case <-ctx.Done():
		// The context was cancelled, either due to a timeout,
		// or any other cancellation cause.
		// Remove the channel from the list.
		// If Signal or Broadcast has just been called, the channel may
		// have disappeared, i.e. notification and cancellations have occurred
		// at almost the same time.
		c.internalLock.Lock()
		defer c.internalLock.Unlock()
		for i, ch := range c.waitChans {
			if ch == newWaitChan {
				close(ch)
				c.waitChans = c.waitChans[i+1:]
				break
			}
		}
		return false
	case <-newWaitChan:
		// The goroutine was awakened either by
		// a call to Signal or a call to Broadcast.
		c.L.Lock()
		return true
	}
}
