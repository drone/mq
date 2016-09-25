package server

import (
	"bytes"
	"errors"

	"github.com/drone/mq/stomp"
)

// ErrNotAuthorized is returned when the peer connection is not
// authorized to establish a connection with the STOMP server.
var ErrNotAuthorized = errors.New("stomp: not authorized")

// Authorizer is a callback function used to authenticate a peer
// connection prior to establishing the session. If the callback
// returns a non-nil error an error message is sent to the peer
// and the connection is closed.
type Authorizer func(*stomp.Message) error

// BasicAuth is a authorization callback function that authorizes
// the peer connection using a basic, global username and password.
func BasicAuth(username, password string) Authorizer {
	var (
		user = []byte(username)
		pass = []byte(password)
	)
	return func(m *stomp.Message) (err error) {
		if bytes.Equal(m.User, user) && bytes.Equal(m.Pass, pass) {
			return nil
		}
		return ErrNotAuthorized
	}
}
