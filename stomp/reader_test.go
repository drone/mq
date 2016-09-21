package stomp

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestRead(t *testing.T) {
	for _, test := range payloads {
		message := NewMessage()
		err := message.Parse([]byte(test.payload))
		if err != nil {
			t.Errorf("error parsing message %q", test.payload)
			continue
		}

		if !reflect.DeepEqual(test.message, message) {
			t.Errorf("problems parsing message %q", test.payload)
			pretty.Ldiff(t, test.message, message)
		}
	}
}

// these are all malformed and should result in an error
// and should not cause the parser to blow up.
func TestReadMalformed(t *testing.T) {
	var tests = []string{
		"",                       // no header
		"STOMP",                  // no header newline
		"STOMP\nversion",         // no header separator
		"STOMP\nversion:",        // no header value
		"STOMP\nversion:1.1.2",   // no header newline
		"STOMP\nversion:1.1.2\n", // no newline before eof
	}

	for _, test := range tests {
		message := NewMessage()
		err := message.Parse([]byte(test))
		if err == nil {
			t.Errorf("Want error parsing message %q", test)
		}
	}
}

var resultmsg *Message

func BenchmarkParse(b *testing.B) {
	var msg *Message
	var err error

	msg = NewMessage()
	defer msg.Release()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		msg.Reset()

		err = msg.Parse(sampleMessage)
		if err != nil {
			b.Error(err)
			return
		}
	}
	resultmsg = msg
}

var sampleMessage = []byte(`PUBLISH
version:1.0.0
destination:/topic/test

foo
bar
baz
qux`)
