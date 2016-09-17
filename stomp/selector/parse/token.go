package parse

// Token is a lexical token.
type Token uint

// List of lexical tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF

	// Identifiers and basic type literals
	IDENT
	TEXT
	REAL
	INTEGER

	// Operators and delimiters
	EQ  // ==
	LT  // <
	LTE // <=
	GT  // >
	GTE // >=
	NEQ // !=

	COMMA  // ,
	LPAREN // (
	LBRACK // [

	RPAREN // )
	RBRACK // ]

	// Keywords
	NOT
	AND
	OR
	IN
	GLOB
	REGEXP
	TRUE
	FALSE
)

// String returns the string representation of a token.
func (t Token) String() string {
	return tokenString[t]
}

var tokenString = map[Token]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	IDENT:   "IDENT",
	TEXT:    "TEXT",
	REAL:    "REAL",
	INTEGER: "INTEGER",
	EQ:      "EQ",
	LT:      "LT",
	LTE:     "LTE",
	GT:      "GT",
	GTE:     "GTE",
	NEQ:     "NEQ",
	COMMA:   "COMMA",
	LPAREN:  "LPAREN",
	LBRACK:  "LBRACK",
	RPAREN:  "RPAREN",
	RBRACK:  "RBRACK",
	NOT:     "NOT",
	AND:     "AND",
	OR:      "OR",
	IN:      "IN",
	GLOB:    "GLOB",
	REGEXP:  "REGEXP",
	TRUE:    "TRUE",
	FALSE:   "FALSE",
}
