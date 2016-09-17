package parse

import (
	"unicode"
	"unicode/utf8"
)

// scanner implements a lexical scanner that reads unicode characters
// and tokens from a raw buffer.
type scanner struct {
	buf   []byte
	pos   int
	start int
	width int
}

// scan reads the next token or Unicode character from source and
// returns it. It returns EOF at the end of the source.
func (s *scanner) scan() Token {
	s.start = s.pos
	s.skipWhitespace()

	r := s.read()
	switch {
	case isIdent(r):
		s.unread()
		return s.scanIdent()
	case isQuote(r):
		s.unread()
		return s.scanQuote()
	case isNumeric(r):
		s.unread()
		return s.scanNumber()
	case isCompare(r):
		s.unread()
		return s.scanCompare()
	}

	switch r {
	case eof:
		return EOF
	case '(':
		return LPAREN
	case ')':
		return RPAREN
	case ',':
		return COMMA
	}

	return ILLEGAL
}

// bytes returns the bytes corresponding to the most recently scanned
// token. Valid after calling Scan().
func (s *scanner) bytes() []byte {
	return s.buf[s.start:s.pos]
}

// init initializes a scanner with a new buffer and returns s.
func (s *scanner) init(buf []byte) {
	s.buf = buf
	s.pos = 0
	s.start = 0
	s.width = 0
}

func (s *scanner) scanIdent() Token {
	for {
		if r := s.read(); r == eof {
			break
		} else if !isAlphaNumeric(r) && r != '_' {
			s.unread()
			break
		}
	}

	ident := s.bytes()
	switch string(ident) {
	case "NOT", "not":
		return NOT
	case "AND", "and":
		return AND
	case "OR", "or":
		return OR
	case "IN", "in":
		return IN
	case "GLOB", "glob":
		return GLOB
	case "REGEXP", "regexp":
		return REGEXP
	case "TRUE", "true":
		return TRUE
	case "FALSE", "false":
		return FALSE
	}

	return IDENT
}

func (s *scanner) scanQuote() (tok Token) {
	s.read() // consume first quote

	for {
		if r := s.read(); r == eof {
			return ILLEGAL
		} else if isQuote(r) {
			break
		}
	}
	return TEXT
}

func (s *scanner) scanNumber() Token {
	for {
		if r := s.read(); r == eof {
			break
		} else if !isNumeric(r) {
			s.unread()
			break
		}
	}
	return INTEGER
}

func (s *scanner) scanCompare() (tok Token) {
	switch s.read() {
	case '=':
		tok = EQ
	case '!':
		tok = NEQ
	case '>':
		tok = GT
	case '<':
		tok = LT
	}

	r := s.read()
	switch {
	case tok == GT && r == '=':
		tok = GTE
	case tok == LT && r == '=':
		tok = LTE
	case tok == EQ && r == '=':
		tok = EQ
	case tok == NEQ && r == '=':
		tok = NEQ
	case tok == NEQ && r != '=':
		tok = ILLEGAL
	default:
		s.unread()
	}
	return
}

func (s *scanner) skipWhitespace() {
	for {
		if r := s.read(); r == eof {
			break
		} else if !isWhitespace(r) {
			s.unread()
			break
		}
	}
	s.ignore()
}

func (s *scanner) read() rune {
	if s.pos >= len(s.buf) {
		s.width = 0
		return eof
	}
	r, w := utf8.DecodeRune(s.buf[s.pos:])
	s.width = w
	s.pos += s.width
	return r
}

func (s *scanner) unread() {
	s.pos -= s.width
}

func (s *scanner) ignore() {
	s.start = s.pos
}

// eof rune sent when end of file is reached
var eof = rune(0)

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func isNumeric(r rune) bool {
	return unicode.IsDigit(r) || r == '.'
}

func isAlphaNumeric(r rune) bool {
	return r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isQuote(r rune) bool {
	return r == '\''
}

func isCompare(r rune) bool {
	return r == '=' || r == '!' || r == '>' || r == '<'
}

func isIdent(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '-' || r == '[' || r == ']'
}
