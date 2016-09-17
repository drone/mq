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
	EQL // ==
	LSS // <
	GTR // >

	NEQ // !=
	LEQ // <=
	GEQ // >=

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
