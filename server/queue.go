package server

import (
	"container/list"
	"math/rand"
	"sync"
	"time"

	"github.com/drone/mq/stomp"
)

type queue struct {
	sync.RWMutex

	dest []byte
	subs map[*subscription]struct{}
	list *list.List
}

func newQueue(dest []byte) *queue {
	return &queue{
		dest: dest,
		subs: make(map[*subscription]struct{}),
		list: list.New(),
	}
}

func (q *queue) publish(m *stomp.Message) error {
	c := m.Copy()
	c.ID = stomp.Rand()
	c.Method = stomp.MethodMessage

	q.Lock()
	q.list.PushBack(c)
	q.Unlock()
	return q.process()
}

func (q *queue) subscribe(s *subscription, m *stomp.Message) error {
	q.Lock()
	q.subs[s] = struct{}{}
	q.Unlock()
	return q.process()
}

func (q *queue) unsubscribe(s *subscription, m *stomp.Message) error {
	q.Lock()
	delete(q.subs, s)
	q.Unlock()
	return nil
}

func (q *queue) disconnect(s *session) error {
	q.Lock()
	for _, subscription := range s.sub {
		delete(q.subs, subscription)
	}
	q.Unlock()
	return nil
}

// returns true if the topic has zero subscribers indicating
// that it can be recycled.
func (q *queue) recycle() (ok bool) {
	q.RLock()
	ok = len(q.subs) == 0 && q.list.Len() == 0
	q.RUnlock()
	return
}

// return the destination name.
func (q *queue) destination() string {
	return string(q.dest)
}

func (q *queue) restore(m *stomp.Message) error {
	q.Lock()
	q.list.PushFront(m)
	q.Unlock()
	return q.process()
}

func (q *queue) process() error {
	q.Lock()
	defer q.Unlock()

	var next *list.Element
	for e := q.list.Front(); e != nil; e = next {
		next = e.Next()
		m := e.Value.(*stomp.Message)

		// if the message expires we can remove it from the list
		if len(m.Expires) != 0 && stomp.ParseInt64(m.Expires) < time.Now().Unix() {
			q.list.Remove(e)
			continue
		}

		for _, sub := range shuffle(q.subs) {
			// evaluate against the sql selector
			if sub.selector != nil {
				if ok, _ := sub.selector.Eval(m.Header); !ok {
					continue
				}
			}

			// evaluate against prefetch counts
			if sub.prefetch != 0 && sub.prefetch == sub.pending {
				continue
			}
			// increment the pending prefectch
			if sub.prefetch != 0 {
				sub.pending++
			}
			if sub.ack {
				m.Ack = stomp.Rand()
				sub.session.Lock()
				sub.session.ack[string(m.Ack)] = m.Copy()
				sub.session.Unlock()
			}

			m.Subs = sub.id
			sub.session.send(m)
			q.list.Remove(e)
		}
	}
	return nil
}

// helper function to randomize the list of subscribers in an attempt
// to more evenly distribute messages in a round robin fashion.
//
// NOTE this is a basic implementation and we recognize there is plenty
// of room for improvement here.
func shuffle(subm map[*subscription]struct{}) []*subscription {
	var subs []*subscription
	for sub := range subm {
		subs = append(subs, sub)
	}
	for i := range subs {
		j := rand.Intn(i + 1)
		subs[i], subs[j] = subs[j], subs[i]
	}
	return subs
}
