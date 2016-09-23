package server

import (
	"sync"

	"github.com/drone/mq/stomp/selector"
)

// subscription represents a session subscription to a
// broker topic or queue.
type subscription struct {
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
