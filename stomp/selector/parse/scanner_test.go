package parse

import "testing"

func TestScanner(t *testing.T) {
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

	s := new(scanner)
	s.init(exampleLarge)

	i := 0
	for {
		token := s.scan()
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

func BenchmarkScanner(b *testing.B) {
	s := new(scanner)
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s.init(exampleSmall)

	scanner:
		for {
			result = s.scan()
			if result == EOF {
				break scanner
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
