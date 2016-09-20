package parse

import (
	"unicode"
	"unicode/utf8"
)

// lexer implements a lexical scanner that reads unicode characters
// and tokens from a byte buffer.
type lexer struct {
	buf   []byte
	pos   int
	start int
	width int
}

// scan reads the next token or Unicode character from source and
// returns it. It returns EOF at the end of the source.
func (l *lexer) scan() Token {
	l.start = l.pos
	l.skipWhitespace()

	r := l.read()
	switch {
	case isIdent(r):
		l.unread()
		return l.scanIdent()
	case isQuote(r):
		l.unread()
		return l.scanQuote()
	case isNumeric(r):
		l.unread()
		return l.scanNumber()
	case isCompare(r):
		l.unread()
		return l.scanCompare()
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
func (l *lexer) bytes() []byte {
	return l.buf[l.start:l.pos]
}

// string returns the string corresponding to the most recently scanned
// token. Valid after calling Scan().
func (l *lexer) string() string {
	return string(l.bytes())
}

// init initializes a scanner with a new buffer.
func (l *lexer) init(buf []byte) {
	l.buf = buf
	l.pos = 0
	l.start = 0
	l.width = 0
}

func (l *lexer) scanIdent() Token {
	for {
		if r := l.read(); r == eof {
			break
		} else if !isAlphaNumeric(r) && r != '_' {
			l.unread()
			break
		}
	}

	ident := l.bytes()
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

func (l *lexer) scanQuote() (tok Token) {
	l.read() // consume first quote

	for {
		if r := l.read(); r == eof {
			return ILLEGAL
		} else if isQuote(r) {
			break
		}
	}
	return TEXT
}

func (l *lexer) scanNumber() Token {
	for {
		if r := l.read(); r == eof {
			break
		} else if !isNumeric(r) {
			l.unread()
			break
		}
	}
	return INTEGER
}

func (l *lexer) scanCompare() (tok Token) {
	switch l.read() {
	case '=':
		tok = EQ
	case '!':
		tok = NEQ
	case '>':
		tok = GT
	case '<':
		tok = LT
	}

	r := l.read()
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
		l.unread()
	}
	return
}

func (l *lexer) skipWhitespace() {
	for {
		if r := l.read(); r == eof {
			break
		} else if !isWhitespace(r) {
			l.unread()
			break
		}
	}
	l.ignore()
}

func (l *lexer) read() rune {
	if l.pos >= len(l.buf) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRune(l.buf[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *lexer) unread() {
	l.pos -= l.width
}

func (l *lexer) peek() Token {
	var (
		pos   = l.pos
		start = l.start
		width = l.width
	)
	tok := l.scan()
	l.pos = pos
	l.start = start
	l.width = width
	return tok
}

func (l *lexer) ignore() {
	l.start = l.pos
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
