package stomp

import (
	"bytes"
	"testing"
)

func TestMessageCopy(t *testing.T) {
	m := NewMessage()
	m.ID = 1
	m.Proto = []byte("1.2")
	m.Method = MethodSend
	m.User = []byte("username")
	m.Pass = []byte("password")
	m.Dest = []byte("/topic/test")
	m.Subs = 1
	m.Ack = AckAuto
	m.Prefetch = []byte("2")
	m.Selector = []byte("ram >= 2")
	m.Persist = PersistTrue
	m.Retain = RetainAll
	m.Receipt = []byte("1")
	m.Body = []byte("hello world")
	m.Header.Add([]byte("key"), []byte("val"))

	c := m.Copy()

	if m.ID != c.ID {
		t.Errorf("expect ID value is copied")
	}
	if !bytes.Equal(m.Proto, c.Proto) {
		t.Errorf("expect Proto value is copied")
	}
	if !bytes.Equal(m.Method, c.Method) {
		t.Errorf("expect Method value is copied")
	}
	if !bytes.Equal(m.User, c.User) {
		t.Errorf("expect User value is copied")
	}
	if !bytes.Equal(m.Pass, c.Pass) {
		t.Errorf("expect Pass value is copied")
	}
	if !bytes.Equal(m.Dest, c.Dest) {
		t.Errorf("expect Dest value is copied")
	}
	if m.Subs != c.Subs {
		t.Errorf("expect Subs value is copied")
	}
	if !bytes.Equal(m.Ack, c.Ack) {
		t.Errorf("expect Ack value is copied")
	}
	if !bytes.Equal(m.Prefetch, c.Prefetch) {
		t.Errorf("expect Prefetch value is copied")
	}
	if !bytes.Equal(m.Selector, c.Selector) {
		t.Errorf("expect Selector value is copied")
	}
	if !bytes.Equal(m.Persist, c.Persist) {
		t.Errorf("expect Persist value is copied")
	}
	if !bytes.Equal(m.Retain, c.Retain) {
		t.Errorf("expect Retain value is copied")
	}
	if !bytes.Equal(m.Receipt, c.Receipt) {
		t.Errorf("expect Receipt value is copied")
	}
	if !bytes.Equal(m.Body, c.Body) {
		t.Errorf("expect Body value is copied")
	}
	if m.Header.itemc != c.Header.itemc {
		t.Errorf("expect Header items are copied")
	}
	// for good measure, let's make sure that altering the copy
	// will not alter the original.
	c.Body = []byte("bonjour monde!")
	if bytes.Equal(m.Body, c.Body) {
		t.Errorf("expect updating the Body does not alter original")
	}
	c.Header.items[0].data = []byte("foo")
	if bytes.Equal(m.Header.items[0].data, c.Header.items[0].data) {
		t.Errorf("expect updating the Header does not alter original")
	}
}

func TestMessageRelease(t *testing.T) {
	m := NewMessage()
	m.ID = 1
	m.Proto = []byte("1.2")
	m.Method = MethodSend
	m.User = []byte("username")
	m.Pass = []byte("password")
	m.Dest = []byte("/topic/test")
	m.Subs = 1
	m.Ack = AckAuto
	m.Prefetch = []byte("2")
	m.Selector = []byte("ram >= 2")
	m.Persist = PersistTrue
	m.Retain = RetainAll
	m.Receipt = []byte("1")
	m.Body = []byte("hello world")
	m.Header.Add([]byte("key"), []byte("val"))
	m.Release()

	if m.ID != 0 {
		t.Errorf("expect ID to reset to zero value")
	}
	if len(m.Proto) != 0 {
		t.Errorf("expect Proto to reset to zero value")
	}
	if len(m.Method) != 0 {
		t.Errorf("expect Method to reset to zero value")
	}
	if len(m.User) != 0 {
		t.Errorf("expect User to reset to zero value")
	}
	if len(m.Pass) != 0 {
		t.Errorf("expect Pass to reset to zero value")
	}
	if len(m.Dest) != 0 {
		t.Errorf("expect Dest to reset to zero value")
	}
	if m.Subs != 0 {
		t.Errorf("expect Subs to reset to zero value")
	}
	if len(m.Ack) != 0 {
		t.Errorf("expect Ack to reset to zero value")
	}
	if len(m.Prefetch) != 0 {
		t.Errorf("expect Prefetch to reset to zero value")
	}
	if len(m.Selector) != 0 {
		t.Errorf("expect Selector to reset to zero value")
	}
	if len(m.Persist) != 0 {
		t.Errorf("expect Persist to reset to zero value")
	}
	if len(m.Retain) != 0 {
		t.Errorf("expect Retain to reset to zero value")
	}
	if len(m.Receipt) != 0 {
		t.Errorf("expect Receipt to reset to zero value")
	}
	if len(m.Body) != 0 {
		t.Errorf("expect Body to reset to zero value")
	}
	if m.Header.itemc != 0 {
		t.Errorf("expect Header to reset to zero value")
	}
}
