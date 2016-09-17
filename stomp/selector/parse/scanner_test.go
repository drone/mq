package parse

import "testing"

func TestScanner(t *testing.T) {
	var tokens = []Token{
		IDENT,   // ram
		GEQ,     // >=
		INTEGER, // 2
		AND,     // AND
		LPAREN,  // (
		IDENT,   // private
		EQL,     // ==
		FALSE,   // false
		AND,     // AND
		IDENT,   // admin
		EQL,     // ==
		TRUE,    // TRUE
		RPAREN,  // )
		OR,      // OR
		IDENT,   // org
		IN,      // IN
		LPAREN,  // (
		TEXT,    // drone
		COMMA,   // ,
		TEXT,    //drone-plugins
		RPAREN,  // )
		EOF,     // EOF
	}

	s := NewScanner(data)

	i := 0
	for {
		token := s.Scan()
		if token == EOF {
			break
		}
		if tokens[i] != token {
			t.Errorf("Expected token %v, got %v for %q", tokens[i], token, s.String())
		}
		i++
	}
}

var result Token

func BenchmarkScanner(b *testing.B) {
	s := NewScanner(data)
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		s.Reset(data)

	scanner:
		for {
			result = s.Scan()
			if result == EOF {
				break scanner
			}
		}
	}
}

var data = []byte(`
ram >= 2 AND
(private == false AND admin = TRUE) OR
org IN ('drone', 'drone-plugins')
`)
