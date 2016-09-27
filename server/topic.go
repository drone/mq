package server

import (
	"bytes"
	"sync"

	"github.com/drone/mq/stomp"
)

// topic is a type of destination handler that implements a
// publish subscribe pattern. Subscribers to a topic receive
// all messages from the publisher.
type topic struct {
	sync.RWMutex

	dest []byte
	hist []*stomp.Message
	subs map[*subscription]struct{}
}

func newTopic(dest []byte) *topic {
	return &topic{
		dest: dest,
		subs: make(map[*subscription]struct{}),
	}
}

// publishes a copy of the message to the subsciber list.
// If the message includes the retain:true headers the message is
// saved for future use. If the message includes retain:remove the
// previously retained message is set to nil.
func (t *topic) publish(m *stomp.Message) error {
	id := stomp.Rand()

	t.RLock()
	for sub := range t.subs {
		if sub.selector != nil {
			if ok, _ := sub.selector.Eval(m.Header); !ok {
				continue
			}
		}
		c := m.Copy()
		c.ID = id
		c.Method = stomp.MethodMessage
		c.Subs = sub.id
		sub.session.send(c)
	}
	t.RUnlock()

	// if a message has the retain header set we should either
	// retain the message, or remove the existing retained message.
	if len(m.Retain) != 0 {
		c := m.Copy()

		t.Lock()
		switch {
		case bytes.Equal(m.Retain, stomp.RetainLast):
			if len(t.hist) == 1 {
				t.hist[0] = c
			} else {
				t.hist = t.hist[:0]
				t.hist = append(t.hist, c)
			}
		case bytes.Equal(m.Retain, stomp.RetainAll):
			t.hist = append(t.hist, c)
		case bytes.Equal(m.Retain, stomp.RetainRemove):
			t.hist = t.hist[:0]
		}
		t.Unlock()
	}

	return nil
}

// registers the subscription with the topic broker and
// sends the last retained message, if one exists.
func (t *topic) subscribe(s *subscription, m *stomp.Message) error {
	t.Lock()
	t.subs[s] = struct{}{}
	t.Unlock()

	t.RLock()
	hist := make([]*stomp.Message, len(t.hist))
	copy(hist, t.hist)
	t.RUnlock()

	for _, m := range hist {
		c := m.Copy()
		c.Method = stomp.MethodMessage
		c.Subs = s.id
		c.ID = stomp.Rand()
		s.session.send(c)
	}

	return nil
}

func (t *topic) unsubscribe(s *subscription, m *stomp.Message) error {
	t.Lock()
	delete(t.subs, s)
	t.Unlock()
	return nil
}

func (t *topic) disconnect(s *session) error {
	t.Lock()
	for _, subscription := range s.sub {
		delete(t.subs, subscription)
	}
	t.Unlock()
	return nil
}

func (t *topic) process() error {
	return nil
}

func (t *topic) restore(m *stomp.Message) error {
	return nil
}

// returns true if the topic has zero subscribers indicating
// that it can be recycled.
func (t *topic) recycle() (ok bool) {
	t.RLock()
	ok = len(t.subs) == 0 && len(t.hist) == 0
	t.RUnlock()
	return
}

// return the destination name.
func (t *topic) destination() string {
	return string(t.dest)
}
