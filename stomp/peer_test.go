package stomp

import (
	"io"
	"testing"
)

func TestPeer(t *testing.T) {
	a, b := Pipe()

	sent := NewMessage()
	a.Send(sent)

	recv := <-b.Receive()
	if sent != recv {
		t.Errorf("Sending message to pipe a should be received by pipe b")
	}

	a.Close()
	b.Close()

	if a.Send(nil) != io.EOF {
		t.Errorf("Want error when sending a message to a closed peer")
	}
	if b.Send(nil) != io.EOF {
		t.Errorf("Want error when sending a message to a closed peer")
	}
}
