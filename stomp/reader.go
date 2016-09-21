package stomp

import (
	"bytes"
	"fmt"
)

func read(input []byte, m *Message) (err error) {
	var (
		pos int
		off int
		tot = len(input)
	)

	// parse the stomp message
	for ; ; off++ {
		if off == tot {
			return fmt.Errorf("stomp: invalid method")
		}
		if input[off] == '\n' {
			m.Method = input[pos:off]
			off++
			pos = off
			break
		}
	}

	// parse the stomp headers
	for {
		if off == tot {
			return fmt.Errorf("stomp: unexpected eof")
		}
		if input[off] == '\n' {
			off++
			pos = off
			break
		}

		var (
			name  []byte
			value []byte
		)

	loop:
		// parse each individual header
		for ; ; off++ {
			if off >= tot {
				return fmt.Errorf("stomp: unexpected eof")
			}

			switch input[off] {
			case '\n':
				value = input[pos:off]
				off++
				pos = off
				break loop
			case ':':
				name = input[pos:off]
				off++
				pos = off
			}
		}

		switch {
		case bytes.Equal(name, HeaderAccept):
			m.Proto = value
		case bytes.Equal(name, HeaderAck):
			m.Ack = value
		case bytes.Equal(name, HeaderDest):
			m.Dest = value
		case bytes.Equal(name, HeaderExpires):
			m.Expires = parseInt64(value)
		case bytes.Equal(name, HeaderLogin):
			m.User = value
		case bytes.Equal(name, HeaderPass):
			m.Pass = value
		case bytes.Equal(name, HeaderID):
			m.ID = parseInt64(value)
		case bytes.Equal(name, HeaderMessageID):
			m.ID = parseInt64(value)
		case bytes.Equal(name, HeaderPersist):
			m.Persist = value
		case bytes.Equal(name, HeaderPrefetch):
			m.Prefetch = value
		case bytes.Equal(name, HeaderReceipt):
			m.Receipt = value
		case bytes.Equal(name, HeaderReceiptID):
			m.Receipt = value
		case bytes.Equal(name, HeaderRetain):
			m.Retain = value
		case bytes.Equal(name, HeaderSelector):
			m.Selector = value
		case bytes.Equal(name, HeaderSubscription):
			m.Subs = parseInt64(value)
		case bytes.Equal(name, HeaderVersion):
			m.Proto = value
		default:
			m.Header.Add(name, value)
		}
	}

	if tot > pos {
		m.Body = input[pos:]
	}
	return
}

const (
	asciiZero = 48
	asciiNine = 57
)

// parseInt64 returns the ascii integer value.
func parseInt64(d []byte) (n int64) {
	if len(d) == 0 {
		return 0
	}
	for _, dec := range d {
		if dec < asciiZero || dec > asciiNine {
			return 0
		}
		n = n*10 + (int64(dec) - asciiZero)
	}
	return n
}
