package server

import (
	"sync"

	"github.com/drone/mq/stomp/selector"
)

// subscription represents a session subscription to a
// broker topic or queue.
type subscription struct {
	mu sync.Mutex

	id       []byte
	dest     []byte
	ack      bool
	prefetch int
	pending  int
	session  *session
	selector *selector.Selector
}

// reset the subscription properties to zero values.
func (s *subscription) reset() {
	s.id = s.id[:0]
	s.dest = s.dest[:0]
	s.ack = false
	s.prefetch = 0
	s.pending = 0
	s.session = nil
	s.selector = nil
}

// release releases the subscription to the pool.
func (s *subscription) release() {
	s.reset()
	subscriptionPool.Put(s)
}

// Pending returns the pending message count.
func (s *subscription) Pending() (i int) {
	s.mu.Lock()
	i = s.pending
	s.mu.Unlock()
	return
}

// PendingIncr increments the pending message count.
func (s *subscription) PendingIncr() {
	s.mu.Lock()
	s.pending++
	s.mu.Unlock()
}

// PendingDecr decrements the pending message count.
func (s *subscription) PendingDecr() {
	s.mu.Lock()
	if s.pending != 0 {
		s.pending--
	}
	s.mu.Unlock()
}

//
// subscription pool
//

var subscriptionPool = sync.Pool{New: createSubscription}

func createSubscription() interface{} {
	return &subscription{}
}

func requestSubscription() *subscription {
	return subscriptionPool.Get().(*subscription)
}
