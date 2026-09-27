package audit

import (
	"testing"
)

func TestWalkPaths(t *testing.T) {
	root := map[string]any{
		"services": []any{map[string]any{"debug": true}},
		"a.b":      map[string]any{"": 1},
	}

	var paths []string
	err := Walk(Node{Value: root},
		func(n Node) { paths = append(paths, n.Path.String()) })
	if err != nil {
		t.Fatal(err)
	}
}

func TestWalkArrayContext(t *testing.T) {
	root := map[string]any{"algorithms": []any{"MD5", "SHA256"}}

	count := 0
	err := Walk(Node{Value: root}, func(n Node) {
		if _, ok := n.Value.(string); ok {
			count++
			if n.Key != "algorithms" || n.Parent == nil {
				t.Fatalf("losed array context: %+v", n)
			}
		}
	})
	if err != nil || count != 2 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestPathEscapesControlCharacters(t *testing.T) {
	path := Path{{Key: "line\nsecret"}, {IsIndex: true, Index: 2}}
	if got, want := path.String(), `$["line\nsecret"][2]`; got != want {
		t.Fatalf("%q != %q", got, want)
	}
}
