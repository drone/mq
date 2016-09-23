package server

import (
	"bytes"
	"testing"

	"github.com/drone/mq/stomp"
)

func Test_topic_publish(t *testing.T) {
	m := stomp.NewMessage()
	m.Dest = []byte("/topic/test")
	m.Body = []byte("hello")
	m.Selector = []byte("skip != true")
	defer m.Release()

	peer, client := stomp.Pipe()
	sess := requestSession()
	sess.peer = peer
	defer sess.release()

	s := sess.subs(m)
	b := newTopic(m.Dest)
	b.subscribe(s, m)
	b.publish(m)

	select {
	case got := <-client.Receive():
		if !bytes.Equal(got.Body, m.Body) {
			t.Errorf("expect message published to topic subscription")
		}
	default:
		t.Errorf("expect message published and delivered")
	}

	skip := stomp.NewMessage()
	skip.Header.Add([]byte("skip"), []byte("true"))
	skip.Body = []byte("skip me")
	b.publish(skip)
	select {
	case <-client.Receive():
		t.Errorf("expect the selector to filter out the message.")
	default:
		// expected
	}
}

func Test_topic_publish_retain(t *testing.T) {
	m := stomp.NewMessage()
	m.Dest = []byte("/topic/test")
	m.Body = []byte("hello")
	m.Retain = stomp.RetainLast
	defer m.Release()

	b := newTopic(m.Dest)
	b.publish(m)
	if len(b.hist) != 1 || !bytes.Equal(b.hist[0].Body, m.Body) {
		t.Errorf("expected topic retained message")
	}

	m.Retain = stomp.RetainLast
	m.Body = []byte("hello2")
	b.publish(m)
	if len(b.hist) != 1 || !bytes.Equal(b.hist[0].Body, m.Body) {
		t.Errorf("expected topic retained message to update")
	}

	m.Retain = stomp.RetainAll
	m.Body = []byte("hello")
	b.publish(m)
	if len(b.hist) != 2 {
		t.Errorf("expected topic retained message to append")
	}

	m.Retain = stomp.RetainRemove
	b.publish(m)
	if len(b.hist) != 0 {
		t.Errorf("expected topic retained message removed")
	}
}

func Test_topic_subscribe(t *testing.T) {
	peer, client := stomp.Pipe()
	sess := requestSession()
	sess.peer = peer
	defer sess.release()

	msg1 := stomp.NewMessage()
	msg1.Method = stomp.MethodSend
	msg1.Dest = []byte("/topic/test")
	msg1.Body = []byte("hello")
	msg1.Retain = stomp.RetainAll
	defer msg1.Release()

	msg2 := stomp.NewMessage()
	msg2.Method = stomp.MethodSubscribe
	msg2.Dest = []byte("/topic/test")
	defer msg2.Release()

	msg3 := stomp.NewMessage()
	msg3.Method = stomp.MethodUnsubscribe
	msg3.Dest = []byte("/topic/test")
	defer msg3.Release()

	brok := newTopic(msg1.Dest)
	brok.publish(msg1)

	sub := sess.subs(msg2)
	defer sess.unsub(sub)

	brok.subscribe(sub, msg2)
	if _, ok := brok.subs[sub]; !ok {
		t.Errorf("want subscription added to topic")
	}

	got := <-client.Receive()
	if !bytes.Equal(got.Body, msg1.Body) {
		t.Errorf("want retained message sent to subscriber")
	}

	brok.unsubscribe(sub, msg3)
	if _, ok := brok.subs[sub]; ok {
		t.Errorf("want subscription removed from topic")
	}
}

func Test_topic_disconnect(t *testing.T) {
	sess := requestSession()
	defer sess.release()

	msg := stomp.NewMessage()
	msg.Dest = []byte("/topic/test")
	defer msg.Release()

	sub := sess.subs(msg)
	defer sess.unsub(sub)

	brok := newTopic(msg.Dest)
	brok.subscribe(sub, msg)
	if _, ok := brok.subs[sub]; !ok {
		t.Errorf("want subscription added to topic")
	}

	brok.disconnect(sess)
	if _, ok := brok.subs[sub]; ok {
		t.Errorf("want subscription removed from topic on disconnect")
	}
}

func Test_topic_recycle(t *testing.T) {
	dest := []byte("/topic/test")
	brok := newTopic(dest)
	if !brok.recycle() {
		t.Errorf("want recycle true when no subscribers")
	}
	msg := stomp.NewMessage()
	defer msg.Release()

	brok.hist = []*stomp.Message{msg}
	if brok.recycle() {
		t.Errorf("want recycle false when no subscribers but retained message")
	}
	brok.hist = brok.hist[:0]
	brok.subs[&subscription{}] = struct{}{}
	if brok.recycle() {
		t.Errorf("want recycle false when subscribers")
	}
}

func Test_topic_dest(t *testing.T) {
	dest := []byte("/topic/test")
	brok := newTopic(dest)
	if got := brok.destination(); got != "/topic/test" {
		t.Errorf("want destingation name /topic/test got %s", got)
	}
}
