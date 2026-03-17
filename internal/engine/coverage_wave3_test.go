package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestArrayLazyParsePathMoreBranches(t *testing.T) {
	t.Run("partial parse reaches requested index", func(t *testing.T) {
		node, err := Parse([]byte(`[1,{"a":2},3]`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		arr := node.(*arrayNode)
		arr.lazyParsePath([]string{"1"})
		if len(arr.value) != 2 || arr.parsed.Load() {
			t.Fatalf("expected partial parse only, len=%d parsed=%v", len(arr.value), arr.parsed.Load())
		}
		if got := arr.value[1].Get("a").Int(); got != 2 {
			t.Fatalf("unexpected parsed child value: %d", got)
		}
	})

	t.Run("empty array exits cleanly", func(t *testing.T) {
		node, err := Parse([]byte(`[]`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		arr := node.(*arrayNode)
		arr.lazyParsePath([]string{"0"})
		if !arr.parsed.Load() {
			t.Fatal("expected empty array path parse to mark parsed")
		}
		if got := arr.Index(0); got.IsValid() {
			t.Fatal("expected empty array index to stay invalid")
		}
	})

	t.Run("child parse error is stored", func(t *testing.T) {
		arr := NewArrayNode(nil, []byte(`[truX]`), nil).(*arrayNode)
		arr.lazyParsePath([]string{"0"})
		if arr.Error() == nil {
			t.Fatal("expected malformed child parse error")
		}
	})
}

func TestBaseHelperLookupAndDirtyBranches(t *testing.T) {
	obj, err := Parse([]byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if key, ok := findObjectChildKey(obj.(*objectNode), NewNumberNode(nil, []byte("1"), nil)); ok || key != "" {
		t.Fatalf("expected missing object child lookup, got %q %v", key, ok)
	}

	arrNode, err := Parse([]byte(`[1,2]`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if idx, ok := findArrayChildIndex(arrNode.(*arrayNode), NewNumberNode(nil, []byte("1"), nil)); ok || idx != 0 {
		t.Fatalf("expected missing array child lookup, got %d %v", idx, ok)
	}

	root, err := MustParse([]byte(`{"outer":[{"leaf":1}]}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	outer := root.Get("outer").(*arrayNode)
	inner := outer.Index(0).(*objectNode)
	markAncestorNodesDirty(inner)
	if !inner.isDirty || !outer.isDirty || !root.(*objectNode).isDirty {
		t.Fatal("expected markAncestorNodesDirty to mark full parent chain")
	}
	markAncestorNodesDirty(NewStringNode(nil, "x", nil))
}

func TestTryFastSlashQueryMoreBranches(t *testing.T) {
	parsed, err := MustParse([]byte(`{"a":{"b":1},"x":2}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	if got := tryFastSlashQuery(parsed, "/a/b"); !got.IsValid() || got.Int() != 1 {
		t.Fatalf("expected parsed fast slash path to work, got %v err=%v", got.Interface(), got.Error())
	}
	if got := tryFastSlashQuery(parsed, "/missing"); got.IsValid() {
		t.Fatal("expected missing parsed fast slash path to be invalid")
	}

	root, err := Parse([]byte(`{"a":{"b":3}}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	obj := root.(*objectNode)
	obj.value = map[string]core.Node{"a": NewObjectNode(obj, []byte(`{"b":3}`), nil)}
	if got := tryFastSlashQuery(obj, "/a/b"); !got.IsValid() || got.Int() != 3 {
		t.Fatalf("expected cached child fast slash path to work, got %v err=%v", got.Interface(), got.Error())
	}

	if got := tryFastSlashQuery(NewObjectNode(nil, []byte(`{"a" 1}`), nil), "/a"); got != nil {
		t.Fatal("expected malformed raw object to abort fast slash query")
	}
	if got := tryFastSlashQuery(NewObjectNode(nil, nil, nil), "/a"); got != nil {
		t.Fatal("expected empty raw object to return nil in fast slash query")
	}
	if got := tryFastSlashQuery(NewArrayNode(nil, []byte(`[1]`), nil), "/a"); got != nil {
		t.Fatal("expected non-object fast slash start to return nil")
	}
}

func TestTryRawDirectPathMoreBranches(t *testing.T) {
	root, err := Parse([]byte(`{"a\"b":{"items":[true,null,{"name":"ok"}]}}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if got := tryRawDirectPath(root, ""); got != nil {
		t.Fatal("expected empty raw direct path to be rejected")
	}
	if got := tryRawDirectPath(root, `/a"b/items[0]`); !got.IsValid() || got.Type() != core.Bool || !got.Bool() {
		t.Fatalf("expected bool raw direct path result, got type=%v val=%v err=%v", got.Type(), got.Bool(), got.Error())
	}
	if got := tryRawDirectPath(root, `/a"b/items[1]`); !got.IsValid() || got.Type() != core.Null {
		t.Fatalf("expected null raw direct path result, got type=%v err=%v", got.Type(), got.Error())
	}
	if got := tryRawDirectPath(root, `/a"b/items[2]/name`); !got.IsValid() || got.String() != "ok" {
		t.Fatalf("expected nested object raw direct path result, got %q err=%v", got.String(), got.Error())
	}

	arr, err := Parse([]byte(`[ {"x":1}, {"x":2} ]`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if got := tryRawDirectPath(arr, `[1]/x`); !got.IsValid() || got.Int() != 2 {
		t.Fatalf("expected array-root raw direct path result, got %v err=%v", got.Interface(), got.Error())
	}

	if got := tryRawDirectPath(root, `/a"b/items[x]`); got != nil {
		t.Fatal("expected malformed array index syntax to return nil")
	}
	if got := tryRawDirectPath(NewObjectNode(nil, []byte(`{"a" 1}`), nil), "/a"); got == nil || got.IsValid() {
		t.Fatal("expected malformed raw object to produce invalid raw direct path result")
	}
	if got := tryRawDirectPath(NewObjectNode(nil, nil, nil), "/a"); got != nil {
		t.Fatal("expected empty raw object to abort raw direct path")
	}
}

func TestInvalidNodeForEachNoop(t *testing.T) {
	called := false
	sharedInvalidNode().ForEach(func(keyOrIndex interface{}, value core.Node) {
		called = true
	})
	if called {
		t.Fatal("expected invalid node ForEach to be a no-op")
	}
}