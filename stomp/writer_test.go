package stomp

import (
	"bytes"
	"testing"
)

var payloads = []struct {
	message *Message
	payload string
}{
	{
		payload: "STOMP\naccept-version:1.2\nlogin:janedoe\npasscode:pa55word\n\n",
		message: &Message{
			Method: MethodStomp,
			Proto:  STOMP,
			User:   []byte("janedoe"),
			Pass:   []byte("pa55word"),
			Header: newHeader(),
		},
	},
	{
		payload: "CONNECTED\nversion:1.2\n\n",
		message: &Message{
			Method: MethodConnected,
			Proto:  STOMP,
			Header: newHeader(),
		},
	},
	{
		payload: "SEND\ndestination:/queue/test\nexpires:1234\nretain:all\npersist:true\nreceipt:4321\n\nhello",
		message: &Message{
			Method:  MethodSend,
			Dest:    []byte("/queue/test"),
			Expires: []byte("1234"),
			Retain:  RetainAll,
			Persist: PersistTrue,
			Receipt: []byte("4321"),
			Body:    []byte("hello"),
			Header:  newHeader(),
		},
	},
	{
		payload: "SUBSCRIBE\nid:123\ndestination:/queue/test\nselector:foo == bar\nprefetch-count:2\nack:auto\n\n",
		message: &Message{
			Method:   MethodSubscribe,
			ID:       []byte("123"),
			Dest:     []byte("/queue/test"),
			Selector: []byte("foo == bar"),
			Prefetch: []byte("2"),
			Ack:      AckAuto,
			Header:   newHeader(),
		},
	},
	{
		payload: "UNSUBSCRIBE\nid:123\n\n",
		message: &Message{
			Method: MethodUnsubscribe,
			ID:     []byte("123"),
			Header: newHeader(),
		},
	},
	{
		payload: "ACK\nid:123\n\n",
		message: &Message{
			Method: MethodAck,
			ID:     []byte("123"),
			Header: newHeader(),
		},
	},
	{
		payload: "NACK\nid:123\n\n",
		message: &Message{
			Method: MethodNack,
			ID:     []byte("123"),
			Header: newHeader(),
		},
	},
	{
		payload: "MESSAGE\nmessage-id:123\ndestination:/queue/test\nsubscription:321\nack:312\n\nhello",
		message: &Message{
			Method: MethodMessage,
			Dest:   []byte("/queue/test"),
			ID:     []byte("123"),
			Subs:   []byte("321"),
			Ack:    []byte("312"),
			Body:   []byte("hello"),
			Header: newHeader(),
		},
	},
	{
		payload: "ERROR\n\n",
		message: &Message{
			Method: MethodError,
			Header: newHeader(),
		},
	},
	{
		payload: "RECEIPT\nreceipt-id:123\n\n",
		message: &Message{
			Method:  MethodRecipet,
			Receipt: []byte("123"),
			Header:  newHeader(),
		},
	},
	{
		payload: "RECEIPT\nreceipt-id:123\nfoo:bar\n\n",
		message: &Message{
			Method:  MethodRecipet,
			Receipt: []byte("123"),
			Header: func() *Header {
				header := newHeader()
				header.Add([]byte("foo"), []byte("bar"))
				return header
			}(),
		},
	},
	{
		payload: "DISCONNECT\nreceipt:123\n\n",
		message: &Message{
			Method:  MethodDisconnect,
			Receipt: []byte("123"),
			Header:  newHeader(),
		},
	},
}

func TestWrite(t *testing.T) {
	for _, test := range payloads {
		if payload := test.message.String(); payload != test.payload {
			t.Errorf("Want serialized message %q, got %q", test.payload, payload)
		}
	}
}

var resultbuf bytes.Buffer

func BenchmarkWrite(b *testing.B) {
	msg := NewMessage()
	msg.Method = MethodSend
	msg.Dest = []byte("/queue/foo")
	msg.Body = []byte("foo\nbar\nbaz\nqux")
	defer msg.Release()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		resultbuf.Reset()
		writeTo(&resultbuf, msg)
	}
}
