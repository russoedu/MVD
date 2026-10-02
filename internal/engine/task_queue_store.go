package engine

import (
	"context"
	"sync"
)

// taskQueue is an unbounded FIFO of entry ids with blocking Pop.
type taskQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []int
	closed bool
}

func newTaskQueue() *taskQueue {
	q := &taskQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *taskQueue) Push(id int) {
	q.mu.Lock()
	q.items = append(q.items, id)
	q.mu.Unlock()
	q.cond.Signal()
}

// Pop blocks until an item is available, the queue is closed or ctx ends.
func (q *taskQueue) Pop(ctx context.Context) (int, bool) {
	stop := context.AfterFunc(ctx, func() {
		q.mu.Lock()
		q.mu.Unlock()
		q.cond.Broadcast()
	})
	defer stop()

	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.closed && ctx.Err() == nil {
		q.cond.Wait()
	}
	if len(q.items) == 0 {
		return 0, false
	}
	id := q.items[0]
	q.items = q.items[1:]
	return id, true
}

func (q *taskQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.cond.Broadcast()
}
