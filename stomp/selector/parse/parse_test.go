package parse

import (
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func TestParse(t *testing.T) {
	tests := []struct {
		query string
		root  BoolExpr
	}{
		{
			query: "ram > 1",
			root: &ComparisonExpr{
				Operator: OperatorGt,
				Left:     &Field{Name: []byte("ram")},
				Right:    &BasicLit{Value: []byte("1")},
			},
		},
		{
			query: "ram >= 1",
			root: &ComparisonExpr{
				Operator: OperatorGte,
				Left:     &Field{Name: []byte("ram")},
				Right:    &BasicLit{Value: []byte("1")},
			},
		},
		{
			query: "ram < 4",
			root: &ComparisonExpr{
				Operator: OperatorLt,
				Left:     &Field{Name: []byte("ram")},
				Right:    &BasicLit{Value: []byte("4")},
			},
		},
		{
			query: "ram <= 4",
			root: &ComparisonExpr{
				Operator: OperatorLte,
				Left:     &Field{Name: []byte("ram")},
				Right:    &BasicLit{Value: []byte("4")},
			},
		},
		{
			query: "platform == 'linux/amd64'",
			root: &ComparisonExpr{
				Operator: OperatorEq,
				Left:     &Field{Name: []byte("platform")},
				Right:    &BasicLit{Value: []byte("linux/amd64")},
			},
		},
		{
			query: "platform != 'linux/amd64'",
			root: &ComparisonExpr{
				Operator: OperatorNeq,
				Left:     &Field{Name: []byte("platform")},
				Right:    &BasicLit{Value: []byte("linux/amd64")},
			},
		},
		{
			query: "platform GLOB 'linux/*'",
			root: &ComparisonExpr{
				Operator: OperatorGlob,
				Left:     &Field{Name: []byte("platform")},
				Right:    &BasicLit{Value: []byte("linux/*")},
			},
		},

		{
			query: "platform NOT GLOB 'linux/*'",
			root: &ComparisonExpr{
				Operator: OperatorNotGlob,
				Left:     &Field{Name: []byte("platform")},
				Right:    &BasicLit{Value: []byte("linux/*")},
			},
		},
		{
			query: "platform REGEXP 'linux/(.+)'",
			root: &ComparisonExpr{
				Operator: OperatorRe,
				Left:     &Field{Name: []byte("platform")},
				Right:    &BasicLit{Value: []byte("linux/(.+)")},
			},
		},
		{
			query: "platform NOT REGEXP 'linux/(.+)'",
			root: &ComparisonExpr{
				Operator: OperatorNotRe,
				Left:     &Field{Name: []byte("platform")},
				Right:    &BasicLit{Value: []byte("linux/(.+)")},
			},
		},
		{
			query: "platform IN ('linux/amd64', 'linux/arm')",
			root: &ComparisonExpr{
				Operator: OperatorIn,
				Left:     &Field{Name: []byte("platform")},
				Right: &ArrayLit{
					Values: []ValExpr{
						&BasicLit{Value: []byte("linux/amd64")},
						&BasicLit{Value: []byte("linux/arm")},
					},
				},
			},
		},
		{
			query: "platform NOT IN ('linux/amd64', 'linux/arm')",
			root: &ComparisonExpr{
				Operator: OperatorNotIn,
				Left:     &Field{Name: []byte("platform")},
				Right: &ArrayLit{
					Values: []ValExpr{
						&BasicLit{Value: []byte("linux/amd64")},
						&BasicLit{Value: []byte("linux/arm")},
					},
				},
			},
		},
		{
			query: "ram > 1 AND cpu >= 2",
			root: &AndExpr{
				Left: &ComparisonExpr{
					Operator: OperatorGt,
					Left:     &Field{Name: []byte("ram")},
					Right:    &BasicLit{Value: []byte("1")},
				},
				Right: &ComparisonExpr{
					Operator: OperatorGte,
					Left:     &Field{Name: []byte("cpu")},
					Right:    &BasicLit{Value: []byte("2")},
				},
			},
		},
		{
			query: "ram > 1 OR cpu >= 2",
			root: &OrExpr{
				Left: &ComparisonExpr{
					Operator: OperatorGt,
					Left:     &Field{Name: []byte("ram")},
					Right:    &BasicLit{Value: []byte("1")},
				},
				Right: &ComparisonExpr{
					Operator: OperatorGte,
					Left:     &Field{Name: []byte("cpu")},
					Right:    &BasicLit{Value: []byte("2")},
				},
			},
		},
		{
			query: "NOT ram < 2",
			root: &NotExpr{
				Expr: &ComparisonExpr{
					Operator: OperatorLt,
					Left:     &Field{Name: []byte("ram")},
					Right:    &BasicLit{Value: []byte("2")},
				},
			},
		},
		{
			query: "ram > 1 AND NOT cpu <= 2",
			root: &AndExpr{
				Left: &ComparisonExpr{
					Operator: OperatorGt,
					Left:     &Field{Name: []byte("ram")},
					Right:    &BasicLit{Value: []byte("1")},
				},
				Right: &NotExpr{
					Expr: &ComparisonExpr{
						Operator: OperatorLte,
						Left:     &Field{Name: []byte("cpu")},
						Right:    &BasicLit{Value: []byte("2")},
					},
				},
			},
		},
	}

	for _, want := range tests {
		buf := []byte(want.query)
		got, err := Parse(buf)
		if err != nil {
			t.Error(err)
			continue
		}
		if !reflect.DeepEqual(got.Root, want.root) {
			t.Errorf("got.Root does not match expected want.root for %q", want.query)
			pretty.Ldiff(t, got.Root, want.root)
		}
	}
}

// Test error conditions.
func TestParseErrors(t *testing.T) {
	tests := []struct {
		query string
		error string
	}{
		{"platform == 'linux/amd64", "selector: parse error:12: illegal value expression"},
		{"platform IN 'linux/amd64'", "selector: parse error:12: unexpected token, expecting ("},
		{"platform IN ('linux/amd64'", "selector: parse error:13: unexpected eof, expecting )"},
		{"platform && 'linux/amd64'", "selector: parse error:9: illegal operator"},
	}

	for _, test := range tests {
		buf := []byte(test.query)
		_, err := Parse(buf)
		if err == nil {
			t.Errorf("expect error parsing %q", test.query)
			continue
		}
		if err.Error() != test.error {
			t.Errorf("expect error %q parsing %q, got %q", test.error, test.query, err)
		}
	}
}
