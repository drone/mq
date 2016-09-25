package server

import (
	"bytes"
	"testing"

	"github.com/drone/mq/stomp"
)

func TestAck(t *testing.T) {
	client, server := stomp.Pipe()

	// this message will subscribe to a queue with
	// client acknowledgements enabled.
	sub := stomp.NewMessage()
	sub.Dest = []byte("/queue/test")
	sub.Ack = stomp.AckClient
	sess := requestSession()
	sess.peer = server

	msg := stomp.NewMessage()
	msg.Dest = []byte("/queue/test")
	msg.Body = []byte("bonjour")

	router := newRouter()
	router.subscribe(sess, sub)
	router.publish(msg)

	got := <-client.Receive()
	if !bytes.Equal(msg.Body, got.Body) {
		t.Errorf("Expect message received by subscriber")
	}
	if _, ok := sess.ack[string(got.Ack)]; !ok {
		t.Errorf("Expect message ack pending for subscriber")
	}
	ack := stomp.NewMessage()
	ack.ID = got.Ack
	router.ack(sess, ack)
	if _, ok := sess.ack[string(got.Ack)]; ok {
		t.Errorf("Expect message ack processed")
	}
}

func TestAckDisconnect(t *testing.T) {
	client, server := stomp.Pipe()

	// this message will subscribe to a queue with
	// client acknowledgements enabled.
	sub := stomp.NewMessage()
	sub.Dest = []byte("/queue/test")
	sub.Ack = stomp.AckClient
	sess := requestSession()
	sess.peer = server

	msg := stomp.NewMessage()
	msg.Dest = []byte("/queue/test")
	msg.Body = []byte("bonjour!")

	router := newRouter()
	router.publish(msg)

	queue := router.destinations["/queue/test"].(*queue)
	// verify the queue has a single item
	if got := queue.list.Len(); got != 1 {
		t.Errorf("Expect queue has 1 message enqueued. Got %d", got)
	}

	router.subscribe(sess, sub)
	// verify the queue has 1 subscriber
	if got := len(queue.subs); got != 1 {
		t.Errorf("Expect queue has 1 subscriber. Got %d", got)
	}

	got := <-client.Receive()
	if !bytes.Equal(msg.Body, got.Body) {
		t.Errorf("Expect message received by subscriber")
	}
	if got := len(sess.ack); got != 1 {
		t.Errorf("Expect message ack count 1, got %d", got)
	}
	if _, ok := sess.ack[string(got.Ack)]; !ok {
		t.Errorf("Expect message ack pending for subscriber")
	}

	// verify the queue is empty after popping the item
	if got := queue.list.Len(); got != 0 {
		t.Errorf("Expect message received and queue empty. Got %d", got)
	}

	router.disconnect(sess)
	if got := len(sess.ack); got != 0 {
		t.Errorf("Expect message ack removed. %d pending acks", got)
	}

	// the queue should have the message re-added
	if queue.list.Len() == 1 {
		t.Errorf("Expect message re-added to the queue")
	}
}
