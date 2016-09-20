package parse

import (
	"encoding/json"
	"testing"
)

func TestParse(t *testing.T) {
	tree, err := Parse([]byte("ram > 2 AND platform == 'linux/amd64' OR foo IN (a, b, c)"))
	if err != nil {
		t.Error(err)
	}

	out, _ := json.MarshalIndent(tree, " ", " ")
	println(string(out))
}
