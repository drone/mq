package server

import (
	"sync"

	"github.com/drone/mq/stomp/selector"
)

// subscription represents a session subscription to a
// broker topic or queue.
type subscription struct {
	id       int64
	dest     []byte
	ack      bool
	prefetch int
	pending  int
	session  *session
	selector *selector.Selector
}

// reset the subscription properties to zero values.
func (s *subscription) reset() {
	s.id = 0
	s.dest = s.dest[:0]
	s.ack = false
	s.prefetch = 0
	s.pending = 0
	s.session = nil
	s.selector = nil
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

func releaseSubscription(s *subscription) {
	s.reset()
	subscriptionPool.Put(s)
}

func seedSubscriptions(count int) {
	for i := 0; i < count; i++ {
		subscriptionPool.Put(
			createSubscription(),
		)
	}
}
