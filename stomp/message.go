package stomp

import (
	"bytes"
	"math/rand"
	"strconv"
	"sync"
)

// Message represents a parsed STOMP message.
type Message struct {
	ID       []byte // id header
	Proto    []byte // stomp version
	Method   []byte // stomp method
	User     []byte // username header
	Pass     []byte // password header
	Dest     []byte // destination header
	Subs     []byte // subscription id
	Ack      []byte // ack id
	Msg      []byte // message-id header
	Persist  []byte // persist header
	Retain   []byte // retain header
	Prefetch []byte // prefetch count
	Expires  []byte // expires header
	Receipt  []byte // receipt header
	Selector []byte // selector header
	Body     []byte
	Header   *Header // custom headers
}

// Copy returns a copy of the Message.
func (m *Message) Copy() *Message {
	c := NewMessage()
	c.ID = m.ID
	c.Proto = m.Proto
	c.Method = m.Method
	c.User = m.User
	c.Pass = m.Pass
	c.Dest = m.Dest
	c.Subs = m.Subs
	c.Ack = m.Ack
	c.Prefetch = m.Prefetch
	c.Selector = m.Selector
	c.Persist = m.Persist
	c.Retain = m.Retain
	c.Receipt = m.Receipt
	c.Expires = m.Expires
	c.Body = m.Body
	c.Header.itemc = m.Header.itemc
	copy(c.Header.items, m.Header.items)
	return c
}

// Apply applies the options to the message.
func (m *Message) Apply(opts ...MessageOption) {
	for _, opt := range opts {
		opt(m)
	}
}

// Parse parses the raw bytes into the message.
func (m *Message) Parse(b []byte) error {
	return read(b, m)
}

// Bytes returns the Message in raw byte format.
func (m *Message) Bytes() []byte {
	var buf bytes.Buffer
	writeTo(&buf, m)
	return buf.Bytes()
}

// String returns the Message in string format.
func (m *Message) String() string {
	return string(m.Bytes())
}

// Release releases the message back to the message pool.
func (m *Message) Release() {
	m.Reset()
	pool.Put(m)
}

// Reset resets the meesage fields to their zero values.
func (m *Message) Reset() {
	m.ID = m.ID[:0]
	m.Proto = m.Proto[:0]
	m.Method = m.Method[:0]
	m.User = m.User[:0]
	m.Pass = m.Pass[:0]
	m.Dest = m.Dest[:0]
	m.Subs = m.Subs[:0]
	m.Ack = m.Ack[:0]
	m.Prefetch = m.Prefetch[:0]
	m.Selector = m.Selector[:0]
	m.Persist = m.Persist[:0]
	m.Retain = m.Retain[:0]
	m.Receipt = m.Receipt[:0]
	m.Expires = m.Expires[:0]
	m.Body = m.Body[:0]
	m.Header.reset()
}

// NewMessage returns an empty message from the message pool.
func NewMessage() *Message {
	return pool.Get().(*Message)
}

var pool = sync.Pool{New: func() interface{} {
	return &Message{Header: newHeader()}
}}

// Rand returns a random int64 number as a []byte of
// ascii characters.
func Rand() []byte {
	return strconv.AppendInt(nil, rand.Int63(), 10)
}
