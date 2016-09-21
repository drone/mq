package parse

import (
	"bytes"
	"fmt"
)

// Tree is the representation of a single parsed SQL statement.
type Tree struct {
	Root BoolExpr

	// Parsing only; cleared after parse.
	lex *lexer
}

// Parse parses the SQL statement and returns a Tree.
func Parse(buf []byte) (*Tree, error) {
	t := new(Tree)
	t.lex = new(lexer)
	return t.Parse(buf)
}

// Parse parses the SQL statement buffer to construct an ast
// representation for execution.
func (t *Tree) Parse(buf []byte) (tree *Tree, err error) {
	defer t.recover(&err)
	t.lex.init(buf)
	t.Root = t.parseExpr()
	return t, nil
}

// recover is the handler that turns panics into returns.
func (t *Tree) recover(err *error) {
	if e := recover(); e != nil {
		*err = e.(error)
	}
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	format = fmt.Sprintf("selector: parse error:%d: %s", t.lex.start, format)
	panic(fmt.Errorf(format, args...))
}

func (t *Tree) parseExpr() BoolExpr {
	left := t.parseVal()
	node := t.parseComparison(left)

	switch t.lex.scan() {
	case OR:
		return t.parseOr(node)
	case AND:
		return t.parseAnd(node)
	default:
		return node
	}
}

func (t *Tree) parseAnd(left BoolExpr) BoolExpr {
	node := new(AndExpr)
	node.Left = left
	node.Right = t.parseExpr()
	return node
}

func (t *Tree) parseOr(left BoolExpr) BoolExpr {
	node := new(OrExpr)
	node.Left = left
	node.Right = t.parseExpr()
	return node
}

func (t *Tree) parseComparison(left ValExpr) BoolExpr {
	var negate bool
	if t.lex.peek() == NOT {
		t.lex.scan()
		negate = true
	}

	node := new(ComparisonExpr)
	node.Operator = t.parseOperator()
	node.Left = left

	if negate {
		switch node.Operator {
		case OperatorIn:
			node.Operator = OperatorNotIn
		case OperatorGlob:
			node.Operator = OperatorNotGlob
		case OperatorRe:
			node.Operator = OperatorNotRe
		}
	}

	switch node.Operator {
	case OperatorIn, OperatorNotIn:
		node.Right = t.parseList()
	case OperatorRe, OperatorNotRe:
		node.Right = t.parseVal()
	default:
		node.Right = t.parseVal()
	}
	return node
}

func (t *Tree) parseOperator() (op Operator) {
	switch t.lex.scan() {
	case EQ:
		return OperatorEq
	case GT:
		return OperatorGt
	case GTE:
		return OperatorGte
	case LT:
		return OperatorLt
	case LTE:
		return OperatorLte
	case NEQ:
		return OperatorNeq
	case IN:
		return OperatorIn
	case REGEXP:
		return OperatorRe
	case GLOB:
		return OperatorGlob
	default:
		t.errorf("illegal operator")
		return
	}
}

func (t *Tree) parseVal() ValExpr {
	switch t.lex.scan() {
	case IDENT:
		node := new(Field)
		node.Name = t.lex.bytes()
		return node
	case TEXT:
		return t.parseText()
	case REAL, INTEGER, TRUE, FALSE:
		node := new(BasicLit)
		node.Value = t.lex.bytes()
		return node
	default:
		t.errorf("illegal value expression")
		return nil
	}
}

func (t *Tree) parseList() ValExpr {
	if t.lex.scan() != LPAREN {
		t.errorf("unexpected token, expecting (")
		return nil
	}
	node := new(ArrayLit)
	for {
		next := t.lex.peek()
		switch next {
		case EOF:
			t.errorf("unexpected eof, expecting )")
		case COMMA:
			t.lex.scan()
		case RPAREN:
			t.lex.scan()
			return node
		default:
			child := t.parseVal()
			node.Values = append(node.Values, child)
		}
	}
}

func (t *Tree) parseText() ValExpr {
	node := new(BasicLit)
	node.Value = t.lex.bytes()

	// this is where we strip the starting and ending quote
	// and unescape the string. On the surface this might look
	// like it is subject to index out of bounds errors but
	// it is safe because it is already verified by the lexer.
	node.Value = node.Value[1 : len(node.Value)-1]
	node.Value = bytes.Replace(node.Value, quoteEscaped, quoteUnescaped, -1)
	return node
}

// errString indicates the string literal does no have the right syntax.
// var errString = errors.New("invalid string literal")

var (
	quoteEscaped   = []byte("\\'")
	quoteUnescaped = []byte("'")
)

// unquote interprets buf as a single-quoted literal, returning the
// value that buf quotes.
// func unquote(buf []byte) ([]byte, error) {
// 	n := len(buf)
// 	if n < 2 {
// 		return nil, errString
// 	}
// 	quote := buf[0]
// 	if quote != quoteUnescaped[0] {
// 		return nil, errString
// 	}
// 	if quote != buf[n-1] {
// 		return nil, errString
// 	}
// 	buf = buf[1 : n-1]
// 	return bytes.Replace(buf, quoteEscaped, quoteUnescaped, -1), nil
// }
