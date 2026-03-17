package core

import "testing"

func TestNodeTypeString(t *testing.T) {
	cases := map[NodeType]string{
		Invalid:      "invalid",
		Object:       "object",
		Array:        "array",
		String:       "string",
		Number:       "number",
		Bool:         "bool",
		Null:         "null",
		NodeType(99): "invalid",
	}
	for value, want := range cases {
		if got := value.String(); got != want {
			t.Fatalf("NodeType(%d).String() = %q, want %q", value, got, want)
		}
	}
}