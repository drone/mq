package parse

import "testing"

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
