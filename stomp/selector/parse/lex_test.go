package parse

import "testing"

func TestLexer_scan(t *testing.T) {
	tests := []struct {
		query string
		value string
		token token
	}{
		// scantokenIdent
		{"foo", "foo", tokenIdent},
		{" foo", "foo", tokenIdent},
		{"123", "123", tokenInteger},
		{"'foo'", "'foo'", tokenText},
		// scanCompare
		{">", ">", tokenGt},
		{">=", ">=", tokenGte},
		{"<", "<", tokenLt},
		{"<=", "<=", tokenLte},
		{"!=", "!=", tokenNeq},
		{"=", "=", tokenEq},
		{"==", "==", tokenEq},
		{"!>", "!>", tokenIllegal},
		// scanQuote
		{"'foo'", "'foo'", tokenText},
		{"'bar' ", "'bar'", tokenText},
		{"'baz", "'baz", tokenIllegal},
		// scantokenIdent
		{"foo", "foo", tokenIdent},
		{"foo ", "foo", tokenIdent},
		{"NOT", "NOT", tokenNot},
		{"AND", "AND", tokenAnd},
		{"OR", "OR", tokenOr},
		{"IN", "IN", tokenIn},
		{"GLOB", "GLOB", tokenGlob},
		{"REGEXP", "REGEXP", tokenRegexp},
		{"TRUE", "TRUE", tokenTrue},
		{"FALSE", "FALSE", tokenFalse},
		// scanNumber
		{"1", "1", tokenInteger},
		{"1234 ", "1234", tokenInteger},
		// other
		{"(", "(", tokenLparen},
		{")", ")", tokenRparen},
		{",", ",", tokenComma},
		{"", "", tokenEOF},
		{"~", "~", tokenIllegal},
	}

	for _, test := range tests {
		lex := new(lexer)
		lex.init([]byte(test.query))
		tok := lex.scan()
		if tok != test.token {
			t.Errorf("expected token %v got %v", test.token, tok)
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

	if lex.peek() != tokenGt {
		t.Errorf("expected peek to return greater than token")
	}
	if lex.string() != "ram" {
		t.Errorf("expected peek to return tokenIdent token")
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
	var tokens = []token{
		tokenIdent,   // ram
		tokenGte,     // >=
		tokenInteger, // 2
		tokenAnd,     // AND
		tokenLparen,  // (
		tokenIdent,   // private
		tokenEq,      // ==
		tokenFalse,   // false
		tokenAnd,     // AND
		tokenIdent,   // admin
		tokenEq,      // ==
		tokenTrue,    // TRUE
		tokenRparen,  // )
		tokenOr,      // OR
		tokenIdent,   // org
		tokenIn,      // IN
		tokenLparen,  // (
		tokenText,    // drone
		tokenComma,   // ,
		tokenText,    // drone-plugins
		tokenRparen,  // )
		tokenEOF,     // EOF
	}

	lex := new(lexer)
	lex.init(exampleLarge)

	i := 0
	for {
		tok := lex.scan()
		if tok == tokenEOF {
			break
		}
		if tokens[i] != tok {
			t.Errorf("Expected token %v, got %v", tokens[i], tok)
		}
		i++
	}
}

var result token

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
			if result == tokenEOF {
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
