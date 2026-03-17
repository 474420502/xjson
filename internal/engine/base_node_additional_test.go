package engine

import (
	"fmt"
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestBaseNodeHelpersAndFormatting(t *testing.T) {
	root, err := MustParse([]byte(`{"obj":{"plain":1,"special.key":2},"arr":[{"n":1},{"n":2}],"value":"2024-01-02T03:04:05Z"}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}

	if path := root.Query("/obj/plain").Path(); path != "/obj/plain" {
		t.Fatalf("unexpected path: %q", path)
	}
	if path := root.Query("/obj/['special.key']").Path(); path != "/obj/['special.key']" {
		t.Fatalf("unexpected quoted path: %q", path)
	}
	if path := root.Query("/arr[1]/n").Path(); path != "/arr[1]/n" {
		t.Fatalf("unexpected array path: %q", path)
	}

	if !root.Query("/value").Time().Equal(root.Query("/value").MustTime()) {
		t.Fatal("expected Time and MustTime to agree")
	}
	if _, ok := root.Query("/obj/plain").RawFloat(); !ok {
		t.Fatal("expected RawFloat on number")
	}
	if s, ok := root.Query("/value").RawString(); !ok || s != "2024-01-02T03:04:05Z" {
		t.Fatalf("unexpected RawString: %q ok=%v", s, ok)
	}
	if !root.Query("/value").Contains("2024-01-02T03:04:05Z") {
		t.Fatal("expected Contains to match exact string")
	}
	if len(root.Query("/value").Strings()) != 1 {
		t.Fatal("expected scalar Strings fallback")
	}

	if formatPathKey("") != "['']" || formatPathKey("abc") != "abc" || formatPathKey("a'b") != "['a\\'b']" {
		t.Fatal("unexpected formatPathKey results")
	}
	if !isSimplePathKey("abc1") || isSimplePathKey("1abc") || isSimplePathKey("a-b") {
		t.Fatal("unexpected isSimplePathKey results")
	}
	if got := escapeQuotedPathKey("a\\'b"); got != "a\\\\\\'b" {
		t.Fatalf("unexpected escaped path key: %q", got)
	}

	parent := NewObjectNode(nil, nil, nil).(*objectNode)
	child := NewStringNode(parent, "x", nil)
	parent.value = map[string]core.Node{"k": child}
	if key, ok := findObjectChildKey(parent, child); !ok || key != "k" {
		t.Fatalf("findObjectChildKey failed: %q %v", key, ok)
	}
	arr := NewArrayNode(nil, nil, nil).(*arrayNode)
	arr.value = []core.Node{child}
	if idx, ok := findArrayChildIndex(arr, child); !ok || idx != 0 {
		t.Fatalf("findArrayChildIndex failed: %d %v", idx, ok)
	}
}

func TestBaseNodeSetValueSetByPathAndRawBounds(t *testing.T) {
	root, err := MustParse([]byte(`{"obj":{"n":1},"arr":[1,2]}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}

	replacement := root.Query("/obj/n").SetValue(9)
	if !replacement.IsValid() || root.Query("/obj/n").Int() != 9 {
		t.Fatalf("SetValue on object child failed: %v", replacement.Error())
	}
	replacement = root.Query("/arr[0]").SetValue(7)
	if !replacement.IsValid() || root.Query("/arr[0]").Int() != 7 {
		t.Fatalf("SetValue on array child failed: %v", replacement.Error())
	}
	if node := root.SetByPath("/obj/newkey", "x"); !node.IsValid() || root.Query("/obj/newkey").String() != "x" {
		t.Fatalf("SetByPath key set failed: %v", node.Error())
	}
	if node := root.SetByPath("/arr[1]", 8); !node.IsValid() || root.Query("/arr[1]").Int() != 8 {
		t.Fatalf("SetByPath index set failed: %v", node.Error())
	}
	if bad := root.Query("/obj/n").SetByPath("/unsupported[:]", 1); bad.IsValid() {
		t.Fatal("expected invalid SetByPath for slice op")
	}
	if bad := root.SetByPath("", 1); bad.IsValid() {
		t.Fatal("expected invalid SetByPath for empty path")
	}
	if bad := root.SetByPath("/arr[9]/x", 1); bad.IsValid() {
		t.Fatal("expected invalid SetByPath for out of bounds intermediate index")
	}
	if bad := root.Query("/obj/n").SetByPath("/x/y", 1); bad.IsValid() {
		t.Fatal("expected invalid SetByPath through scalar intermediate node")
	}
	if bad := root.Query("/obj/n").SetByPath("/x", 1); bad.IsValid() {
		t.Fatal("expected invalid SetByPath final key set on scalar")
	}

	var bn baseNode
	if bn.Raw() != "" || bn.RawBytes() != nil {
		t.Fatal("expected zero-value raw helpers to be empty")
	}
	bn.raw = []byte("abcdef")
	bn.start = -2
	bn.end = 100
	if got := bn.Raw(); got != "abcdef" {
		t.Fatalf("unexpected clamped Raw: %q", got)
	}
	bn.start = 5
	bn.end = 3
	if got := bn.Raw(); got != "" {
		t.Fatalf("expected empty Raw on inverted bounds, got %q", got)
	}

	missingChild := &baseNode{parent: NewObjectNode(nil, nil, nil)}
	missingChild.self = missingChild
	if got := missingChild.Path(); got != "/?" {
		t.Fatalf("expected unknown object path marker, got %q", got)
	}
}

func TestBaseNodeApplyAndErrorHelpers(t *testing.T) {
	strNode := NewStringNode(nil, "hello", nil)
	if got := strNode.Apply(core.TransformFunc(func(core.Node) interface{} { return 99 })); !got.IsValid() || got.Int() != 99 {
		t.Fatalf("unexpected Apply transform result: %v", got.Interface())
	}

	base := &baseNode{}
	base.self = base
	base.setError(fmt.Errorf("boom"))
	base.setError(fmt.Errorf("ignored"))
	if base.Error() == nil || base.Error().Error() != "boom" {
		t.Fatalf("unexpected setError result: %v", base.Error())
	}

	funcsNode := &baseNode{}
	funcsNode.self = funcsNode
	called := false
	funcsNode.RegisterFunc("x", func(n core.Node) core.Node {
		called = true
		return n
	})
	funcsNode.CallFunc("x")
	if !called {
		t.Fatal("expected registered function to be called")
	}
	funcsNode.RemoveFunc("x")
	if got := funcsNode.CallFunc("x"); got.IsValid() {
		t.Fatal("expected missing function call to be invalid")
	}

	defaultBase := &baseNode{}
	defaultBase.self = defaultBase
	if got := defaultBase.Get("k"); got.IsValid() {
		t.Fatal("expected default Get to be invalid")
	}
	if got := defaultBase.Index(0); got.IsValid() {
		t.Fatal("expected default Index to be invalid")
	}
	if got := defaultBase.Set("k", 1); got.IsValid() {
		t.Fatal("expected default Set to be invalid")
	}
	if got := defaultBase.Append(1); got.IsValid() {
		t.Fatal("expected default Append to be invalid")
	}
	if got := defaultBase.Filter(func(core.Node) bool { return true }); got.IsValid() {
		t.Fatal("expected default Filter to be invalid")
	}
	if got := defaultBase.Map(func(core.Node) interface{} { return 1 }); got.IsValid() {
		t.Fatal("expected default Map to be invalid")
	}
}
