package util

import (
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go-kit/pkg/util/collection"
	"k8s.io/utils/clock"
)

// adapted from k8s.io/client-go@v0.22.2/util/workqueue/delaying_queue.go

type executableFunc func()

// waitFor holds the executee to add and the time it should be executed
type waitFor struct {
	function executableFunc
	readyAt  time.Time
}

func waitForComparator(first, second *waitFor) bool {
	return first.readyAt.Before(second.readyAt)
}

type DelayingExecutor struct {
	// waitingForAddCh is a buffered channel that feeds waitingForAdd
	waitingForAddCh          chan *waitFor
	clock                    clock.Clock
	stopCh                   chan struct{}
	slowStopCh               chan struct{}
	priorityQueue            collection.PrioritySet[*waitFor]
	closeStopChOnce          sync.Once
	closeSlowStopChOnce      sync.Once
	closeWaitingForAddChOnce sync.Once
}

func NewDelayingExecutor(size int) *DelayingExecutor {
	priorityQueue := collection.NewPriorityQueue[*waitFor](waitForComparator,
		func(first, second *waitFor) bool {
			return first.readyAt == second.readyAt &&
				// Can't simply use `first.function == second.function`,
				// which will encounter "func can only be compared to nil"
				reflect.ValueOf(first).Pointer() != reflect.ValueOf(second).Pointer()
		})

	executor := &DelayingExecutor{
		// Don't need to close the channel, or we may get "panic: send on closed channel"
		waitingForAddCh: make(chan *waitFor, size),
		clock:           clock.RealClock{},
		stopCh:          make(chan struct{}),
		slowStopCh:      make(chan struct{}),
		priorityQueue:   priorityQueue,
	}

	go executor.waitingLoop()
	return executor
}

func (d *DelayingExecutor) ExcuteAfter(f func(), duration time.Duration) {
	runtimeErr := runtimeError("Executor has been shutted down!")
	defer func() {
		if err := recover(); err != nil {
			plainErr, isPlainError := err.(runtime.Error)
			if isPlainError {
				if plainErr.Error() == "send on closed channel" {
					panic(runtimeErr)
				}
			}

			panic(err)
		}
	}()

	select {
	case <-d.stopCh:
		panic(runtimeErr)
	default:
		d.waitingForAddCh <- &waitFor{function: f, readyAt: d.clock.Now().Add(duration)}
	}
}

func (d *DelayingExecutor) waitingLoop() {
	// Make a placeholder channel to use when there are no items in our list
	never := make(<-chan time.Time)

	// Make a timer that expires when the item at the head of the waiting list is ready
	var nextReadyAtTimer clock.Timer

	for {
		now := d.clock.Now()
		// Add ready entries
		for d.priorityQueue.Len() > 0 {
			entry := d.priorityQueue.Peek()
			if entry.readyAt.After(now) {
				break
			}

			entry, _ = d.priorityQueue.Pop()
			go d.executeIgnorePanic(entry.function)
		}

		// Set up a wait for the first item's readyAt (if one exists)
		nextReadyAt := never
		if d.priorityQueue.Len() > 0 {
			if nextReadyAtTimer != nil {
				nextReadyAtTimer.Stop()
			}
			entry := d.priorityQueue.Peek()
			nextReadyAtTimer = d.clock.NewTimer(entry.readyAt.Sub(now))
			nextReadyAt = nextReadyAtTimer.C()
		}

		select {
		case <-d.stopCh:
			return
		case <-nextReadyAt:
		case waitEntry := <-d.waitingForAddCh:
			if waitEntry == nil { // d.waitingForAddCh is closed
				d.drainPriorityQueue()
				d.closeSlowStopChOnce.Do(func() {
					// Can't use d.stopCh here, because in drainPriorityQueue, we have the following snippet,
					// which prevent the remaining tasks from being executed
					// select {
					// case <-d.stopCh:
					// 	return
					// 	...
					// }
					close(d.slowStopCh)
				})
				return
			}
			if waitEntry.readyAt.After(d.clock.Now()) {
				d.priorityQueue.Add(waitEntry)
			} else {
				go d.executeIgnorePanic(waitEntry.function)
			}

			d.drainWaitingForAddCh()
		}
	}
}

func (d *DelayingExecutor) drainPriorityQueue() {
	for entry, existing := d.priorityQueue.Pop(); existing; entry, existing = d.priorityQueue.Pop() {
		nextReadyAtTimer := d.clock.NewTimer(entry.readyAt.Sub(time.Now()))
		select {
		case <-nextReadyAtTimer.C():
			nextReadyAtTimer.Stop()
			go d.executeIgnorePanic(entry.function)
		}
	}
}

func (d *DelayingExecutor) drainWaitingForAddCh() {
	for {
		select {
		case <-d.stopCh:
			return
		case waitEntry := <-d.waitingForAddCh:
			if waitEntry == nil { // d.waitingForAddCh is closed
				return
			}
			if waitEntry.readyAt.After(d.clock.Now()) {
				d.priorityQueue.Add(waitEntry)
			} else {
				go d.executeIgnorePanic(waitEntry.function)
			}
		default:
			return
		}
	}
}

func (d *DelayingExecutor) executeIgnorePanic(executableFunc func()) {
	select {
	case <-d.stopCh:
		return
	default:
		defer func() {
			if r := recover(); r != nil {

			}
		}()

		executableFunc()
	}
}

func (d *DelayingExecutor) ShutDownFast() {
	d.closeStopChOnce.Do(func() { // In case of "close of closed channel"
		close(d.stopCh)
	})

	d.closeSlowStopChOnce.Do(func() {
		close(d.slowStopCh) // Don't want to block ShutDownWithDrain
	})
}

// ShutDownWithDrain This method will reject new tasks immediately, but blocks until after all tasks are guaranteed to
// be executed eventually.
// it can only guarantee that all tasks will be executed eventually after it returns.
// After the return, some tasks may not have finished and some tasks may not even begin.
func (d *DelayingExecutor) ShutDownWithDrain(block bool) {
	d.closeWaitingForAddChOnce.Do(func() {
		// To to make sure after ShutDownWithDrain no tasks will be added to it thread-safely,
		// we can either close d.waitingForAddCh  or use a lock and a flag(let's say `isShutingDownWithDrain bool`)
		close(d.waitingForAddCh)
	})
	if block {
		<-d.slowStopCh
	}
}

type DelayingChannel[T any] struct {
	executor       *DelayingExecutor
	ch             chan T
	isClosed       bool
	closedLock     sync.Locker
	remainingTasks int64
}

func NewDelayingChannel[T any](size int) *DelayingChannel[T] {
	return &DelayingChannel[T]{
		executor:       NewDelayingExecutor(size),
		ch:             make(chan T, size),
		isClosed:       false,
		closedLock:     &sync.Mutex{},
		remainingTasks: 0,
	}
}

func (d *DelayingChannel[T]) Get() T {
	return <-d.ch
}

func (d *DelayingChannel[T]) AddAfter(entry T, duration time.Duration) {
	atomic.AddInt64(&d.remainingTasks, 1)
	d.executor.ExcuteAfter(func() {
		d.ch <- entry
		atomic.AddInt64(&d.remainingTasks, -1)
	}, duration)
}

func (d *DelayingChannel[T]) Close() {
	d.closedLock.Lock()
	defer d.closedLock.Unlock()

	if d.isClosed {
		panic(runtimeError("\"send on closed channel\""))
	}

	// make sure after this method returns, no more tasks can be added
	d.executor.ShutDownWithDrain(false)

	go func() {
		d.executor.ShutDownWithDrain(true)
		// Even ShutDownWithDrain blocks,
		// it can only guarantee that all tasks will be executed eventually after it returns.
		// After the return, some tasks may not have finished and some tasks may not even begin.
		for atomic.LoadInt64(&d.remainingTasks) > 0 {
			time.Sleep(50 * time.Millisecond)
		}
		close(d.ch)
	}()
	d.isClosed = true
}

type runtimeError string

func (r *runtimeError) RuntimeError() {}

func (r *runtimeError) Error() string {
	return string(*r)
}
