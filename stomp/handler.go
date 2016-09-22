package stomp

// Handler handles a STOMP message.
type Handler interface {
	Handle(*Message)
}

// The HandlerFunc type is an adapter to allow the use of an ordinary
// function as a STOMP message handler.
type HandlerFunc func(*Message)

// Handle calls f(m).
func (f HandlerFunc) Handle(m *Message) { f(m) }
