package stomp

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {

	var tests = []struct {
		message *Message
		payload string
	}{
		{
			payload: "STOMP\r\naccept-version:1.2\r\nlogin:janedoe\r\npasscode:pa55word\r\n\r\n",
			message: &Message{
				Method: MethodStomp,
				Proto:  STOMP,
				User:   []byte("janedoe"),
				Pass:   []byte("pa55word"),
				Header: newHeader(),
			},
		},
		{
			payload: "CONNECTED\r\nversion:1.2\r\n\r\n",
			message: &Message{
				Method: MethodConnected,
				Proto:  STOMP,
				Header: newHeader(),
			},
		},
		{
			payload: "SEND\r\ndestination:/queue/test\r\nexpires:1234\r\nretain:all\r\npersist:true\r\nreceipt-id:4321\r\n\r\nhello",
			message: &Message{
				Method:  MethodSend,
				Dest:    []byte("/queue/test"),
				Expires: 1234,
				Retain:  RetainAll,
				Persist: PersistTrue,
				Receipt: []byte("4321"),
				Body:    []byte("hello"),
				Header:  newHeader(),
			},
		},
		{
			payload: "SUBSCRIBE\r\nid:123\r\ndestination:/queue/test\r\nselector:foo == bar\r\nprefetch-count:2\r\nack:auto\r\n\r\n",
			message: &Message{
				Method:   MethodSubscribe,
				ID:       123,
				Dest:     []byte("/queue/test"),
				Selector: []byte("foo == bar"),
				Prefetch: []byte("2"),
				Ack:      AckAuto,
				Header:   newHeader(),
			},
		},
		{
			payload: "UNSUBSCRIBE\r\nid:123\r\n\r\n",
			message: &Message{
				Method: MethodUnsubscribe,
				ID:     123,
				Header: newHeader(),
			},
		},
		{
			payload: "ACK\r\nid:123\r\n\r\n",
			message: &Message{
				Method: MethodAck,
				ID:     123,
				Header: newHeader(),
			},
		},
		{
			payload: "NACK\r\nid:123\r\n\r\n",
			message: &Message{
				Method: MethodNack,
				ID:     123,
				Header: newHeader(),
			},
		},
		{
			payload: "MESSAGE\r\nmessage-id:123\r\ndestination:/queue/test\r\nsubscription:321\r\nack:312\r\n\r\nhello",
			message: &Message{
				Method: MethodMessage,
				Dest:   []byte("/queue/test"),
				ID:     123,
				Subs:   321,
				Ack:    []byte("312"),
				Body:   []byte("hello"),
				Header: newHeader(),
			},
		},
		{
			payload: "RECEIPT\r\nreceipt-id:123\r\n\r\n",
			message: &Message{
				Method:  MethodRecipet,
				Receipt: []byte("123"),
				Header:  newHeader(),
			},
		},
		{
			payload: "RECEIPT\r\nreceipt-id:123\r\nfoo:bar\r\n\r\n",
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
	}

	for _, test := range tests {
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

var sampleMessage = []byte(`PUBLISH
version:1.0.0
destination:/topic/test

foo
bar
baz
qux`)
