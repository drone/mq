package stomp

import (
	"bufio"
	"io"
	"net"
	"time"
)

// default read and write buffer size.
const bufferSize = 32 * 1024

var (
	never    time.Time
	deadline = time.Second * 5
)

type connPeer struct {
	conn net.Conn
	done chan bool

	incoming chan *Message
	outgoing chan *Message
}

// Conn creates a network-connected peer that reads and writes
// messages using net.Conn c.
func Conn(c net.Conn) Peer {
	p := &connPeer{
		incoming: make(chan *Message),
		outgoing: make(chan *Message),
		done:     make(chan bool),
		conn:     c,
	}

	go p.readInto(p.incoming)
	go p.writeFrom(p.outgoing)
	return p
}

func (c *connPeer) Receive() <-chan *Message {
	return c.incoming
}

func (c *connPeer) Send(message *Message) error {
	select {
	case <-c.done:
		return io.EOF
	default:
		c.outgoing <- message
		return nil
	}
}

// TODO we should gracefully shut down, blocking
// until all pending messages are dispatched.
func (c *connPeer) Close() error {
	return c.close()
}

func (c *connPeer) close() (err error) {
	err = c.conn.Close()
	close(c.done)
	close(c.incoming)
	close(c.outgoing)
	return
}

func (c *connPeer) readInto(messages chan<- *Message) {
	bufc := bufio.NewReaderSize(c.conn, bufferSize)
	for {
		c.conn.SetReadDeadline(time.Now().Add(deadline))
		buf, err := bufc.ReadBytes(0)
		if err != nil {
			break
		}
		c.conn.SetReadDeadline(never)

		msg := NewMessage()
		msg.Parse(buf[:len(buf)-1])

		select {
		case <-c.done:
			return
		default:
			messages <- msg
		}
	}
}

func (c *connPeer) writeFrom(messages <-chan *Message) {
	defer c.Close()

	tick := time.NewTicker(time.Millisecond * 100).C
	bufc := bufio.NewWriterSize(c.conn, bufferSize)
	for {
		select {
		case <-c.done:
			return
		case <-tick:
			c.conn.SetWriteDeadline(time.Now().Add(deadline))
			if err := bufc.Flush(); err != nil {
				return
			}
			c.conn.SetWriteDeadline(never)
		case msg := <-messages:
			writeTo(bufc, msg)
			bufc.WriteByte(0)
			msg.Release()
		}
	}
}
