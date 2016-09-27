package server

import (
	"bytes"
	"errors"
	"sync"

	"github.com/drone/mq/logger"
	"github.com/drone/mq/stomp"
)

var (
	errStompMethod    = errors.New("stomp: expected stomp method")
	errNoSubscription = errors.New("stomp: no such subscription")
	errNoDestination  = errors.New("stomp: no such destination")
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
	process() error
	recycle() bool
}

type router struct {
	sync.RWMutex
	authorizer   Authorizer
	destinations map[string]handler
	sessions     map[*session]struct{}
}

func newRouter() *router {
	return &router{
		destinations: make(map[string]handler),
		sessions:     make(map[*session]struct{}),
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
	sub, ok := sess.sub[string(m.ID)]
	if !ok {
		logger.Noticef("stomp: unsubscribe %s: subscription not found",
			string(m.ID),
		)
		return errNoSubscription
	}
	defer sess.unsub(sub)

	r.Lock()
	h, ok := r.destinations[string(sub.dest)]
	r.Unlock()
	if !ok {
		logger.Noticef("stomp: unsubscribe %s: destination not found: %s",
			string(m.ID),
			string(sub.dest),
		)
		return errNoDestination
	}

	logger.Noticef("stomp: unsubscribe %s: successful: destination %s",
		string(m.ID),
		string(sub.dest),
	)

	defer r.collect(h)
	return h.unsubscribe(sub, m)
}

func (r *router) ack(sess *session, m *stomp.Message) {
	sess.Lock()
	ack, ok := sess.ack[string(m.ID)]
	delete(sess.ack, string(m.ID))
	sess.Unlock()

	if ok {
		logger.Verbosef("stomp: ack %s: successful",
			string(m.ID),
		)
	} else {
		logger.Noticef("stomp: ack %s: message not found",
			string(m.ID),
		)
	}

	// if the subscription is still active, check the prefetch
	// count and decrement pending prefetches.
	// TODO this is probably not threadsafe. need to lock the subscription
	// in the event that sub.prefetch is being accessed at the same time.
	sess.Lock()
	sub, ok := sess.sub[string(ack.Subs)]
	if ok && sub.prefetch != 0 && sub.pending > 0 {
		sub.pending--
	}
	sess.Unlock()

	// if prefetch is enabled for the subscription we should re-process
	// the queue now that the subscription pending ack cound is reduced.
	if ok && sub.prefetch != 0 {
		r.RLock()
		h, ok := r.destinations[string(sub.dest)]
		r.RUnlock()
		if ok {
			h.process()
		}
	}

	// if r.storage != nil {
	// 	r.storage.delete(m)
	// }
}

func (r *router) nack(sess *session, m *stomp.Message) {
	sess.Lock()
	nack, ok := sess.ack[string(m.ID)]
	delete(sess.ack, string(m.ID))

	if ok {
		logger.Verbosef("stomp: nack %s: successful",
			string(m.ID),
		)
	} else {
		logger.Noticef("stomp: nack %s: message not found",
			string(m.ID),
		)
	}

	// if the subscription is still active, check the prefetch
	// count and decrement pending prefetches.
	// TODO this is probably not threadsafe. need to lock the subscription
	sub, subscribed := sess.sub[string(nack.Subs)]
	if subscribed && sub.prefetch != 0 && sub.pending > 0 {
		sub.pending--
	}
	sess.Unlock()

	if ok {
		nack.ID = m.Ack
		nack.Ack = m.Ack[:0]
		r.publish(nack)
	}
}

func (r *router) disconnect(sess *session) {
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

	for _, m := range sess.ack {
		delete(sess.ack, string(m.Ack))

		m.ID = m.Ack
		m.Ack = m.Ack[:0]
		r.publish(m)
	}

	r.Lock()
	delete(r.sessions, sess)
	r.Unlock()
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

	// optional message logging
	logger.Debugf("stomp: received message from client.\n%s", message)

	if r.authorizer != nil {
		err := r.authorizer(message)
		if err != nil {
			return err
		}
	}
	session.init(message)

	r.Lock()
	r.sessions[session] = struct{}{}
	r.Unlock()

	// send CONNECTED message indicating the client connection
	// was accepted by the server.
	connected := stomp.NewMessage()
	connected.Method = stomp.MethodConnected
	connected.Proto = stomp.STOMP
	session.send(connected)

	for {
		message, ok := <-session.peer.Receive()
		if !ok {
			return nil
		}

		// optional message logging
		logger.Debugf("stomp: received message from client.\n%s", message)

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
	case bytes.HasPrefix(m.Dest, routeQueue):
		return newQueue(m.Dest)
	default:
		return newQueue(m.Dest)
	}
}
