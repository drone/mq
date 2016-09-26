package stomp

import (
	"io"
	"net"
	"sync"
)

// Peer defines a peer-to-peer connection.
type Peer interface {
	// Send sends a message.
	Send(*Message) error

	// Receive returns a channel of inbound messages.
	Receive() <-chan *Message

	// Close closes the connection.
	Close() error

	// Addr returns the peer address.
	Addr() string
}

// Pipe creates a synchronous in-memory pipe, where reads on one end are
// matched with writes on the other. This is useful for direct, in-memory
// client-server communication.
func Pipe() (Peer, Peer) {
	atob := make(chan *Message, 10)
	btoa := make(chan *Message, 10)

	a := &localPeer{
		incoming: btoa,
		outgoing: atob,
		finished: make(chan bool),
	}
	b := &localPeer{
		incoming: atob,
		outgoing: btoa,
		finished: make(chan bool),
	}

	return a, b
}

type localPeer struct {
	finished chan bool
	outgoing chan<- *Message
	incoming <-chan *Message
}

func (p *localPeer) Receive() <-chan *Message {
	return p.incoming
}

func (p *localPeer) Send(m *Message) error {
	select {
	case <-p.finished:
		return io.EOF
	default:
		p.outgoing <- m
		return nil
	}
}

func (p *localPeer) Close() error {
	close(p.finished)
	close(p.outgoing)
	return nil
}

func (p *localPeer) Addr() string {
	peerAddrOnce.Do(func() {
		// get the local address list
		addr, _ := net.InterfaceAddrs()
		if len(addr) != 0 {
			// use the last address in the list
			peerAddr = addr[len(addr)-1].String()
		}
	})
	return peerAddr
}

var peerAddrOnce sync.Once

// default address displayed for local pipes
var peerAddr = "127.0.0.1/8"
