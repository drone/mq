package stomp

import "bytes"

const defaultHeaderLen = 5

type item struct {
	name []byte
	data []byte
}

// Header represents the header section of the STOMP message.
type Header struct {
	items []item
	itemc int
}

func newHeader() *Header {
	return &Header{
		items: make([]item, defaultHeaderLen),
	}
}

// Get returns the named header value.
func (h *Header) Get(name []byte) (b []byte) {
	for i := 0; i < h.itemc; i++ {
		if v := h.items[i]; bytes.Equal(v.name, name) {
			return v.data
		}
	}
	return
}

// Field returns the named header value in string format. This is used to
// provide compatibility with the SQL expression evaluation package.
func (h *Header) Field(name []byte) []byte {
	return h.Get(name)
}

// Add appens the key value pair to the header.
func (h *Header) Add(name, data []byte) {
	if h.itemc > defaultHeaderLen-1 {
		h.items = append(h.items, item{})
	}
	h.items[h.itemc].name = name
	h.items[h.itemc].data = data
	h.itemc++
}

func (h *Header) reset() {
	h.itemc = 0
	h.items = h.items[:defaultHeaderLen]
	for i := range h.items {
		h.items[i].name = zeroBytes
		h.items[i].data = zeroBytes
	}
}

var zeroBytes []byte
