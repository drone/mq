package server

import (
	"bytes"
	"strconv"
	"sync"

	"github.com/drone/mq/stomp"
	"github.com/drone/mq/stomp/selector"
)

// session represents a single client session (ie connection)
type session struct {
	peer stomp.Peer

	// id  int64
	seq int64
	sub map[string]*subscription
	ack map[string]*stomp.Message

	sync.Mutex
}

// send writes the message to the transport.
func (s *session) send(m *stomp.Message) {
	s.peer.Send(m)
}

// create a subscription for the current session using the
// subscription settings from the given message.
func (s *session) subs(m *stomp.Message) *subscription {
	sub := requestSubscription()
	sub.id = strconv.AppendInt(nil, s.seq, 10)
	sub.dest = m.Dest
	sub.ack = bytes.Equal(m.Ack, stomp.AckClient) || len(m.Prefetch) != 0
	sub.prefetch = stomp.ParseInt(m.Prefetch)
	sub.session = s

	if len(m.Selector) != 0 {
		// TODO we should parse this somewhere else so we can
		// return an error message to the client
		sub.selector, _ = selector.Parse(m.Selector)
	}

	s.sub[string(sub.id)] = sub
	s.seq++
	return sub
}

// remove the subscription from the session and release
// to the session pool.
func (s *session) unsub(sub *subscription) {
	delete(s.sub, string(sub.id))
	sub.release()
}

// reset the session properties to zero values.
func (s *session) reset() {
	s.peer = nil
	s.seq = 0
	for id := range s.sub {
		delete(s.sub, id)
	}
	for id := range s.ack {
		delete(s.ack, id)
	}
}

// release releases the session to the pool.
func (s *session) release() {
	s.reset()
	sessionPool.Put(s)
}

//
// session pool
//

var sessionPool = sync.Pool{New: createSession}

func createSession() interface{} {
	return &session{
		sub: make(map[string]*subscription),
		ack: make(map[string]*stomp.Message),
	}
}

func requestSession() *session {
	return sessionPool.Get().(*session)
}
