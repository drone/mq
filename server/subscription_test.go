package server

import (
	"testing"
)

func init() {
	seedSubscriptions(10)
}

func Test_subscription_reset(t *testing.T) {
	sub := &subscription{
		id:       1,
		dest:     []byte("/topic/test"),
		ack:      true,
		prefetch: 1,
		pending:  1,
		session:  &session{},
	}
	sub.reset()

	if sub.id != 0 {
		t.Errorf("expect subscription id reset")
	}
	if len(sub.dest) != 0 {
		t.Errorf("expect subscription destination reset")
	}
	if sub.ack != false {
		t.Errorf("expect subscription ack flag reset")
	}
	if sub.prefetch != 0 {
		t.Errorf("expect subscription prefetch count reset")
	}
	if sub.pending != 0 {
		t.Errorf("expect subscription pending cout reset")
	}
	if sub.session != nil {
		t.Errorf("expect session subscription reset")
	}
}

func Test_subscription_pool(t *testing.T) {
	s := requestSubscription()
	if s == nil {
		t.Errorf("expected subscription from pool")
	}
	releaseSubscription(s)
}
