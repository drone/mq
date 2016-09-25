package server

import (
	"testing"

	"github.com/drone/mq/stomp"
)

func TestOptions(t *testing.T) {
	s := NewServer(WithCredentials("janedoe", "password"))
	if s.router.authorizer == nil {
		t.Errorf("Expect WithCredentials configures authorizer")
	}

	m := stomp.NewMessage()
	if s.router.authorizer(m) != ErrNotAuthorized {
		t.Errorf("Expect failed authorization when empty username")
	}

	m.Reset()
	m.User = []byte("janedoe")
	m.Pass = []byte("password")
	if err := s.router.authorizer(m); err != nil {
		t.Errorf("Expect successful authorization, got error %s", err)
	}
}
