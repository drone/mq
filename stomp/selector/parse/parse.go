package parse

import (
	"bytes"
	"fmt"
)

// Tree is the representation of a single parsed SQL statement.
type Tree struct {
	Root BoolExpr
	text string

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
	format = fmt.Sprintf("selector: parse error:%d: %s", t.lex.pos, format)
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
		t.errorf("unexpected operator")
		return
	}
}

func (t *Tree) parseVal() ValExpr {
	switch t.lex.scan() {
	case IDENT:
		node := new(Field)
		node.Name = t.lex.bytes()
		return node
	case TEXT, REAL, INTEGER, TRUE, FALSE:
		node := new(BasicLit)
		node.Value = t.lex.bytes()
		node.Value = unquote(node.Value)
		return node
	default:
		t.errorf("unexpected value")
		return nil
	}
}

func (t *Tree) parseList() ValExpr {
	if t.lex.scan() != LPAREN {
		t.errorf("expecting left paren")
		return nil
	}
	node := new(ArrayLit)
	for {
		next := t.lex.peek()
		switch next {
		case EOF:
			t.errorf("unexpected eof")
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

// simple helper function to unquote a literal.
func unquote(buf []byte) []byte {
	n := len(buf)
	if n < 2 {
		return buf
	}
	quote := buf[0]
	if quote != '\'' {
		return buf
	}
	if quote != buf[n-1] {
		return buf
	}
	buf = buf[1 : n-1]
	return bytes.Replace(buf, []byte("\\'"), []byte("'"), -1)
}
