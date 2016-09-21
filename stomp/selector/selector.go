package selector

import "github.com/drone/mq/stomp/selector/parse"

// Selector reprents a parsed SQL selector statement.
type Selector struct {
	*parse.Tree
}

// Parse parses the SQL statement and returns a new Statement object.
func Parse(b []byte) (selector *Selector, err error) {
	selector = new(Selector)
	selector.Tree, err = parse.Parse(b)
	return
}

// Eval evaluates the SQL statement using the provided data and returns true
// if all conditions are satisfied. If a runtime error is experiences a false
// value is returned along with an error message.
func (s *Selector) Eval(row Row) (match bool, err error) {
	defer errRecover(&err)
	state := &state{vars: row}
	match = state.walk(s.Root)
	return
}

// Row defines a row of columnar data.
//
// Note that the field name and field values are represented as []byte
// since stomp header names and values are represented as []byte to avoid
// extra allocations when converting from []byte to string.
type Row interface {
	Field([]byte) []byte
}
