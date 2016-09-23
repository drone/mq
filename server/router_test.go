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
