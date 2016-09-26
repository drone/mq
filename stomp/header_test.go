package stomp

import (
	"bytes"
	"fmt"
	"testing"
)

func TestHeader(t *testing.T) {
	header := newHeader()
	if got := len(header.items); got != defaultHeaderLen {
		t.Errorf("Want the default items len %v, got %d", defaultHeaderLen, got)
	}

	for i := 0; i < 10; i++ {
		var (
			key  = fmt.Sprintf("col%d", i)
			val  = fmt.Sprintf("dat%d", i)
			keyb = []byte(key)
			valb = []byte(val)
		)

		header.Add(keyb, valb)

		// the default list length is 5 and will be expanded as the
		// list grows. This check verifies the list grows as expected.
		if header.itemc != i+1 {
			t.Errorf("Want header length %d, got %d", i+1, header.itemc)
		}

		// this check verifies header key pairs were added to the list
		// and can can be retrieved by header name
		if got := header.Field(keyb); !bytes.Equal(got, valb) {
			t.Errorf("Want header value %q, got %q", val, string(got))
		}
	}

	// verify header behavior when key does not exist
	if got := header.Get([]byte("foo")); len(got) != 0 {
		t.Errorf("Expect empty slice when key does not exist")
	}

	header.reset()
	// verify the values are all reset properly.
	if header.itemc != 0 {
		t.Errorf("Want the header count reset to zero, got %d", header.itemc)
	}
	if got := len(header.items); got != defaultHeaderLen {
		t.Errorf("Want the reset items len %v, got %d", defaultHeaderLen, got)
	}
	for i := 0; i < defaultHeaderLen; i++ {
		if len(header.items[i].name) != 0 {
			t.Errorf("Want header.items[%d].name reset to the zero value", i)
		}
		if len(header.items[i].data) != 0 {
			t.Errorf("Want header.items[%d].value reset to the zero value", i)
		}
	}

	header.Add([]byte("want-true"), []byte("true"))
	header.Add([]byte("want-false"), []byte("false"))
	if got := header.GetBool("want-true"); !got {
		t.Errorf("Expect header.GetBool parses the boolean value true")
	}
	if got := header.GetBool("want-false"); got {
		t.Errorf("Expect header.GetBool parses the boolean value false")
	}
}
