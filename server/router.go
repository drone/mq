package server

import (
	"bytes"
	"errors"
	"sync"

	"github.com/drone/mq/stomp"
)

var (
	errStompMethod    = errors.New("stop: expected stomp method")
	errNoSubscription = errors.New("stop: no such subscription")
	errNoDestination  = errors.New("stop: no such destination")
)

var (
	routeTopic = []byte("/topic/")
	routeQueue = []byte("/queue/")
)

type handler interface {
	destination() string
	publish(*stomp.Message) error
	restore(*stomp.Message) error
	subscribe(*subscription, *stomp.Message) error
	unsubscribe(*subscription, *stomp.Message) error
	disconnect(*session) error
	recycle() bool
}

type router struct {
	sync.RWMutex
	destinations map[string]handler
}

func newRouter() *router {
	return &router{
		destinations: make(map[string]handler),
	}
}

// publish publishes the message to the brokered destination.
func (r *router) publish(m *stomp.Message) error {
	r.RLock()
	h, ok := r.destinations[string(m.Dest)]
	r.RUnlock()

	if !ok && !shouldCreate(m) {
		return errNoDestination
	}

	// if shouldPersist(m) && r.storage != nil {
	// 	r.storage.put(m)
	// }

	if !ok {
		r.Lock()
		// this duplicate check prevents a possible race condition
		// where the topic didn't exist when we checked above but
		// exists now.
		h, ok = r.destinations[string(m.Dest)]
		if !ok {
			h = createHandler(m)
			r.destinations[string(m.Dest)] = h
		}
		r.Unlock()
	}
	return h.publish(m)
}

// subscribe to the brokered destination.
func (r *router) subscribe(sess *session, m *stomp.Message) (err error) {
	r.Lock()
	h, ok := r.destinations[string(m.Dest)]
	if !ok {
		h = createHandler(m)
		r.destinations[string(m.Dest)] = h
	}
	r.Unlock()
	return h.subscribe(sess.subs(m), m)
}

// unsubscribe from the brokered destination.
func (r *router) unsubscribe(sess *session, m *stomp.Message) (err error) {
	sub, ok := sess.sub[m.ID]
	if ok {
		return errNoSubscription
	}
	defer sess.unsub(sub)

	r.Lock()
	h, ok := r.destinations[string(m.Dest)]
	r.Unlock()
	if !ok {
		return errNoDestination
	}

	defer r.collect(h)
	return h.unsubscribe(sub, m)
}

func (r *router) ack(sess *session, m *stomp.Message) {
	sess.Lock()
	delete(sess.ack, m.ID)
	sess.Unlock()

	// if r.storage != nil {
	// 	r.storage.delete(m)
	// }
}

func (r *router) nack(sess *session, m *stomp.Message) {
	sess.Lock()
	mm, ok := sess.ack[m.ID]
	delete(sess.ack, m.ID)
	sess.Unlock()

	if ok {
		r.publish(mm)
	}
}

func (r *router) disconnect(sess *session) {
	for _, msg := range sess.ack {
		r.publish(msg)
	}
	for _, sub := range sess.sub {
		r.Lock()
		h, ok := r.destinations[string(sub.dest)]
		r.Unlock()
		if !ok {
			continue
		}
		h.disconnect(sess)
		r.collect(h)
	}
}

func (r *router) collect(h handler) {
	r.Lock()
	if h.recycle() {
		delete(r.destinations, h.destination())
	}
	r.Unlock()
}

func (r *router) serve(session *session) error {
	message, ok := <-session.peer.Receive()
	if !ok {
		return nil
	}

	// the first message from the client should be STOMP
	if !bytes.Equal(message.Method, stomp.MethodStomp) {
		return errStompMethod
	}

	// send CONNECTED message indicating the client connection
	// was accepted by the server.
	connected := stomp.NewMessage()
	connected.Method = stomp.MethodConnected
	session.send(connected)

	for {
		message, ok := <-session.peer.Receive()
		if !ok {
			return nil
		}

		switch {
		case bytes.Equal(message.Method, stomp.MethodSend):
			r.publish(message)
		case bytes.Equal(message.Method, stomp.MethodSubscribe):
			r.subscribe(session, message)
		case bytes.Equal(message.Method, stomp.MethodUnsubscribe):
			r.unsubscribe(session, message)
		case bytes.Equal(message.Method, stomp.MethodAck):
			r.ack(session, message)
		case bytes.Equal(message.Method, stomp.MethodNack):
			r.nack(session, message)
		case bytes.Equal(message.Method, stomp.MethodDisconnect):
			message.Release()
			return nil
		}

		if len(message.Receipt) != 0 {
			receipt := stomp.NewMessage()
			receipt.Method = stomp.MethodRecipet
			receipt.Receipt = message.Receipt
			session.send(receipt)
		}
		message.Release()
	}
}

func shouldPersist(m *stomp.Message) bool {
	return len(m.Persist) != 0 && bytes.Equal(m.Persist, stomp.PersistTrue)
}

func shouldCreate(m *stomp.Message) bool {
	return bytes.HasPrefix(m.Dest, routeTopic) == false || len(m.Retain) != 0
}

func createHandler(m *stomp.Message) handler {
	switch {
	case bytes.HasPrefix(m.Dest, routeTopic):
		return newTopic(m.Dest)
	// case bytes.HasPrefix(m.Dest, routeQueue):
	// 	return newQueue(m.Dest)
	default:
		// return newQueue(m.Dest)
		return newTopic(m.Dest)
	}
}
