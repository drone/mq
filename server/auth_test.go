package server

import (
	"testing"

	"github.com/drone/mq/stomp"
)

func TestBasicAuth(t *testing.T) {
	f := BasicAuth("janedoe", "password")

	m := stomp.NewMessage()
	m.Pass = []byte("password")
	if f(m) != ErrNotAuthorized {
		t.Errorf("Expect failed authorization when empty username")
	}

	m.Reset()
	m.User = []byte("johnsmith")
	m.Pass = []byte("password")
	if f(m) != ErrNotAuthorized {
		t.Errorf("Expect failed authorization when invalid username")
	}

	m.Reset()
	m.User = []byte("janedoe")
	if f(m) != ErrNotAuthorized {
		t.Errorf("Expect failed authorization when empty password")
	}

	m.Reset()
	m.User = []byte("janedoe")
	m.Pass = []byte("pa55word")
	if f(m) != ErrNotAuthorized {
		t.Errorf("Expect failed authorization when invalid password")
	}

	m.Reset()
	m.User = []byte("janedoe")
	m.Pass = []byte("password")
	if err := f(m); err != nil {
		t.Errorf("Expect successful authorization, got error %s", err)
	}
}
