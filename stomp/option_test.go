package stomp

import (
	"bytes"
	"testing"
)

func TestOptions(t *testing.T) {
	opt := WithAck("auto")
	msg := NewMessage()
	msg.Apply(opt)
	if !bytes.Equal(msg.Ack, AckAuto) {
		t.Errorf("Want WiathAck to apply ack header")
	}

	opt = WithCredentials("janedoe", "password")
	msg = NewMessage()
	msg.Apply(opt)
	if string(msg.User) != "janedoe" {
		t.Errorf("Want WithCredentials to apply username header")
	}
	if string(msg.Pass) != "password" {
		t.Errorf("Want WithCredentials to apply password header")
	}

	opt = WithExpires(1234)
	msg = NewMessage()
	msg.Apply(opt)
	if !bytes.Equal(msg.Expires, []byte("1234")) {
		t.Errorf("Want WithExpires to apply expires header")
	}

	opt = WithHeader("foo", "bar")
	msg = NewMessage()
	msg.Apply(opt)
	if v := msg.Header.Get([]byte("foo")); string(v) != "bar" {
		t.Errorf("Want WithHeader to add header keypair")
	}

	opt = WithPersistence()
	msg = NewMessage()
	msg.Apply(opt)
	if !bytes.Equal(msg.Persist, PersistTrue) {
		t.Errorf("Want WithPersistence to apply persist header")
	}

	opt = WithPrefetch(2)
	msg = NewMessage()
	msg.Apply(opt)
	if !bytes.Equal(msg.Prefetch, []byte("2")) {
		t.Errorf("Want WithPrefetch to apply persist header")
	}

	opt = WithReceipt()
	msg = NewMessage()
	msg.Apply(opt)
	if len(msg.Receipt) == 0 {
		t.Errorf("Want WithReceipt to apply receipt header")
	}

	opt = WithRetain("last")
	msg = NewMessage()
	msg.Apply(opt)
	if !bytes.Equal(msg.Retain, RetainLast) {
		t.Errorf("Want WithRetain to apply retain header")
	}

	opt = WithSelector("ram > 2")
	msg = NewMessage()
	msg.Apply(opt)
	if !bytes.Equal(msg.Selector, []byte("ram > 2")) {
		t.Errorf("Want WithRetain to apply retain header")
	}
}
