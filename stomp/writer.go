package stomp

import (
	"bytes"
	"io"
)

var (
	crlf       = []byte{'\r', '\n'}
	newline    = []byte{'\n'}
	separator  = []byte{':'}
	terminator = []byte{0}
)

func writeTo(w io.Writer, m *Message) {
	w.Write(m.Method)
	w.Write(newline)

	switch {
	case bytes.Equal(m.Method, MethodStomp):
		// version
		w.Write(HeaderAccept)
		w.Write(separator)
		w.Write(m.Proto)
		w.Write(newline)
		// login
		if len(m.User) != 0 {
			w.Write(HeaderLogin)
			w.Write(separator)
			w.Write(m.User)
			w.Write(newline)
		}
		// passcode
		if len(m.Pass) != 0 {
			w.Write(HeaderPass)
			w.Write(separator)
			w.Write(m.Pass)
			w.Write(newline)
		}
	case bytes.Equal(m.Method, MethodConnected):
		// version
		w.Write(HeaderVersion)
		w.Write(separator)
		w.Write(m.Proto)
		w.Write(newline)
	case bytes.Equal(m.Method, MethodSend):
		// dest
		w.Write(HeaderDest)
		w.Write(separator)
		w.Write(m.Dest)
		w.Write(newline)
		if len(m.Expires) != 0 {
			w.Write(HeaderExpires)
			w.Write(separator)
			w.Write(m.Expires)
			w.Write(newline)
		}
		if len(m.Retain) != 0 {
			w.Write(HeaderRetain)
			w.Write(separator)
			w.Write(m.Retain)
			w.Write(newline)
		}
		if len(m.Persist) != 0 {
			w.Write(HeaderPersist)
			w.Write(separator)
			w.Write(m.Persist)
			w.Write(newline)
		}
	case bytes.Equal(m.Method, MethodSubscribe):
		// id
		w.Write(HeaderID)
		w.Write(separator)
		w.Write(m.ID)
		w.Write(newline)
		// destination
		w.Write(HeaderDest)
		w.Write(separator)
		w.Write(m.Dest)
		w.Write(newline)
		// selector
		if len(m.Selector) != 0 {
			w.Write(HeaderSelector)
			w.Write(separator)
			w.Write(m.Selector)
			w.Write(newline)
		}
		// prefetch
		if len(m.Prefetch) != 0 {
			w.Write(HeaderPrefetch)
			w.Write(separator)
			w.Write(m.Prefetch)
			w.Write(newline)
		}
		if len(m.Ack) != 0 {
			w.Write(HeaderAck)
			w.Write(separator)
			w.Write(m.Ack)
			w.Write(newline)
		}
	case bytes.Equal(m.Method, MethodUnsubscribe):
		// id
		w.Write(HeaderID)
		w.Write(separator)
		w.Write(m.ID)
		w.Write(newline)
	case bytes.Equal(m.Method, MethodAck):
		// id
		w.Write(HeaderID)
		w.Write(separator)
		w.Write(m.ID)
		w.Write(newline)
	case bytes.Equal(m.Method, MethodNack):
		// id
		w.Write(HeaderID)
		w.Write(separator)
		w.Write(m.ID)
		w.Write(newline)
	case bytes.Equal(m.Method, MethodMessage):
		// message-id
		w.Write(HeaderMessageID)
		w.Write(separator)
		w.Write(m.ID)
		w.Write(newline)
		// destination
		w.Write(HeaderDest)
		w.Write(separator)
		w.Write(m.Dest)
		w.Write(newline)
		// subscription
		w.Write(HeaderSubscription)
		w.Write(separator)
		w.Write(m.Subs)
		w.Write(newline)
		// ack
		if len(m.Ack) != 0 {
			w.Write(HeaderAck)
			w.Write(separator)
			w.Write(m.Ack)
			w.Write(newline)
		}
	case bytes.Equal(m.Method, MethodRecipet):
		// receipt-id
		w.Write(HeaderReceiptID)
		w.Write(separator)
		w.Write(m.Receipt)
		w.Write(newline)
	}

	// receipt header
	if includeReceiptHeader(m) {
		w.Write(HeaderReceipt)
		w.Write(separator)
		w.Write(m.Receipt)
		w.Write(newline)
	}

	for i, item := range m.Header.items {
		if m.Header.itemc == i {
			break
		}
		w.Write(item.name)
		w.Write(separator)
		w.Write(item.data)
		w.Write(newline)
	}
	w.Write(newline)
	w.Write(m.Body)
}

func includeReceiptHeader(m *Message) bool {
	return len(m.Receipt) != 0 && !bytes.Equal(m.Method, MethodRecipet)
}
