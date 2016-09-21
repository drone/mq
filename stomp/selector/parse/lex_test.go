package parse

import "testing"

func TestLexer_scan(t *testing.T) {
	tests := []struct {
		query string
		value string
		token Token
	}{
		// scanIdent
		{"foo", "foo", IDENT},
		{" foo", "foo", IDENT},
		{"123", "123", INTEGER},
		{"'foo'", "'foo'", TEXT},
		// scanCompare
		{">", ">", GT},
		{">=", ">=", GTE},
		{"<", "<", LT},
		{"<=", "<=", LTE},
		{"!=", "!=", NEQ},
		{"=", "=", EQ},
		{"==", "==", EQ},
		{"!>", "!>", ILLEGAL},
		// scanQuote
		{"'foo'", "'foo'", TEXT},
		{"'bar' ", "'bar'", TEXT},
		{"'baz", "'baz", ILLEGAL},
		// scanIdent
		{"foo", "foo", IDENT},
		{"foo ", "foo", IDENT},
		{"NOT", "NOT", NOT},
		{"AND", "AND", AND},
		{"OR", "OR", OR},
		{"IN", "IN", IN},
		{"GLOB", "GLOB", GLOB},
		{"REGEXP", "REGEXP", REGEXP},
		{"TRUE", "TRUE", TRUE},
		{"FALSE", "FALSE", FALSE},
		// scanNumber
		{"1", "1", INTEGER},
		{"1234 ", "1234", INTEGER},
		// other
		{"(", "(", LPAREN},
		{")", ")", RPAREN},
		{",", ",", COMMA},
		{"", "", EOF},
		{"~", "~", ILLEGAL},
	}

	for _, test := range tests {
		lex := new(lexer)
		lex.init([]byte(test.query))
		token := lex.scan()
		if token != test.token {
			t.Errorf("expected token %s got %s", test.token, token)
		}
		if value := lex.string(); value != test.value {
			t.Errorf("expected token value %s got %s", test.value, value)
		}
	}
}

func TestLexer_peek(t *testing.T) {
	buf := []byte("ram > 2")
	lex := new(lexer)
	lex.init(buf)
	lex.scan()

	if lex.peek() != GT {
		t.Errorf("expected peek to return greater than token")
	}
	if lex.string() != "ram" {
		t.Errorf("expected peek to return ident token")
	}
}

func TestLexer_skipWhitespace(t *testing.T) {
	buf := []byte("  foo")
	lex := new(lexer)
	lex.init(buf)
	lex.skipWhitespace()
	if lex.start != 2 {
		t.Errorf("expect skipWhitespace advances the starting position")
	}

	empty := []byte("  ")
	lex.init(empty)
	lex.skipWhitespace()
	if lex.start != 2 {
		t.Errorf("expect skipWhitespace exits on eof")
	}
}

func TestLexer_read(t *testing.T) {
	buf := []byte("foo")
	lex := new(lexer)
	lex.init(buf)
	lex.read()

	if lex.read() != 'o' {
		t.Errorf("expected read to increment buffer")
	}
	if lex.pos != 2 {
		t.Errorf("expected read to increment position")
	}
	if lex.width != 1 {
		t.Errorf("expected read to increment width")
	}
	lex.unread()
	if lex.pos != 1 {
		t.Errorf("expected unread to decrement position")
	}
	lex.ignore()
	if lex.start != lex.pos {
		t.Errorf("expected ignore to set the start to the position")
	}

	lex.init(buf)
	lex.read()
	lex.read()
	lex.read()
	if lex.read() != eof {
		t.Errorf("expected read to return eof at end of buffer")
	}
}

// this is a more complex test for the lexer that parses a query
// string with many different tokens, whitespace, new-lines, etc.
func TestLexer(t *testing.T) {
	var tokens = []Token{
		IDENT,   // ram
		GTE,     // >=
		INTEGER, // 2
		AND,     // AND
		LPAREN,  // (
		IDENT,   // private
		EQ,      // ==
		FALSE,   // false
		AND,     // AND
		IDENT,   // admin
		EQ,      // ==
		TRUE,    // TRUE
		RPAREN,  // )
		OR,      // OR
		IDENT,   // org
		IN,      // IN
		LPAREN,  // (
		TEXT,    // drone
		COMMA,   // ,
		TEXT,    // drone-plugins
		RPAREN,  // )
		EOF,     // EOF
	}

	lex := new(lexer)
	lex.init(exampleLarge)

	i := 0
	for {
		token := lex.scan()
		if token == EOF {
			break
		}
		if tokens[i] != token {
			t.Errorf("Expected token %s, got %s", tokens[i], token)
		}
		i++
	}
}

var result Token

// this benmark tests performance and allocations. Note that the lexer
// is currently a zero-allocation lexer. Please ensure that changes to
// have minimal impact to performance and do not add allocations.
func BenchmarkLexer(b *testing.B) {
	lex := new(lexer)
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		lex.init(exampleSmall)

	lexer:
		for {
			result = lex.scan()
			if result == EOF {
				break lexer
			}
		}
	}
}

var exampleSmall = []byte(`ram >= 2 AND platform == 'linux/amd64'`)

var exampleLarge = []byte(`
 ram >= 2 AND
(private == false AND admin = TRUE) OR
org IN ('drone', 'drone-plugins')
`)
