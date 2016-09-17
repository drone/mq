package parse

import (
	"unicode"
	"unicode/utf8"
)

// Scanner implements a lexical scanner that reads Unicode characters
// and tokens from a raw buffer.
type Scanner struct {
	buf   []byte
	pos   int
	start int
	width int
}

// NewScanner returns a new instance of Scanner.
func NewScanner(buf []byte) *Scanner {
	return &Scanner{buf: buf}
}

func (s *Scanner) read() rune {
	if s.pos >= len(s.buf) {
		s.width = 0
		return eof
	}
	r, w := utf8.DecodeRune(s.buf[s.pos:])
	s.width = w
	s.pos += s.width
	return r
}

func (s *Scanner) unread() {
	s.pos -= s.width
}

func (s *Scanner) ignore() {
	s.start = s.pos
}

// Bytes returns the bytes corresponding to the most recently scanned
// token. Valid after calling Scan().
func (s *Scanner) Bytes() []byte {
	return s.buf[s.start:s.pos]
}

// String returns the bytes corresponding to the most recently scanned
// token. Valid after calling Scan().
func (s *Scanner) String() string {
	return string(s.Bytes())
}

// Reset resets the scanner buffer.
func (s *Scanner) Reset(buf []byte) {
	s.buf = buf
	s.pos = 0
	s.start = 0
	s.width = 0
}

// Pos returns the position of the character immediately after the
// character or token returned by the last call to Scan.
func (s *Scanner) Pos() int {
	return s.pos
}

// Scan reads the next token or Unicode character from source and
// returns it. It returns EOF at the end of the source.
func (s *Scanner) Scan() Token {
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

func (s *Scanner) scanIdent() Token {
	for {
		if r := s.read(); r == eof {
			break
		} else if !isAlphaNumeric(r) && r != '_' {
			s.unread()
			break
		}
	}

	ident := s.Bytes()
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

func (s *Scanner) scanQuote() (tok Token) {
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

func (s *Scanner) scanNumber() Token {
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

func (s *Scanner) scanCompare() (tok Token) {
	switch s.read() {
	case '=':
		tok = EQL
	case '!':
		tok = NEQ
	case '>':
		tok = GTR
	case '<':
		tok = LSS
	}

	r := s.read()
	switch {
	case tok == GTR && r == '=':
		tok = GEQ
	case tok == LSS && r == '=':
		tok = LEQ
	case tok == EQL && r == '=':
		tok = EQL
	case tok == NEQ && r == '=':
		tok = NEQ
	case tok == NEQ && r != '=':
		tok = ILLEGAL
	default:
		s.unread()
	}
	return
}

func (s *Scanner) skipWhitespace() {
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
