package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestArrayLazyParseIndexBranches(t *testing.T) {
	t.Run("no raw marks parsed", func(t *testing.T) {
		node := NewArrayNode(nil, nil, nil).(*arrayNode)
		node.lazyParseIndex(0)
		if !node.parsed.Load() {
			t.Fatal("expected parsed to be set when raw is empty")
		}
	})

	t.Run("negative index falls back to full parse", func(t *testing.T) {
		node, err := Parse([]byte(`[1,2,3]`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		arr := node.(*arrayNode)
		arr.lazyParseIndex(-1)
		if !arr.parsed.Load() || len(arr.value) != 3 {
			t.Fatalf("expected full parse for negative index, parsed=%v len=%d", arr.parsed.Load(), len(arr.value))
		}
	})

	t.Run("existing cached index short-circuits", func(t *testing.T) {
		node, err := Parse([]byte(`[1,2,3]`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		arr := node.(*arrayNode)
		arr.value = append(arr.value, NewNumberNode(arr, []byte("1"), nil))
		arr.lazyParseIndex(0)
		if len(arr.value) != 1 {
			t.Fatalf("expected cached value slice to stay size 1, got %d", len(arr.value))
		}
	})

	t.Run("malformed raw falls back", func(t *testing.T) {
		node := NewArrayNode(nil, []byte(`oops`), nil).(*arrayNode)
		node.lazyParseIndex(0)
		if !node.parsed.Load() {
			t.Fatal("expected malformed raw to trigger fallback parse")
		}
	})

	t.Run("partial parse reaches target index", func(t *testing.T) {
		node, err := Parse([]byte(`[1,{"a":2},3]`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		arr := node.(*arrayNode)
		arr.lazyParseIndex(1)
		if len(arr.value) != 2 || arr.parsed.Load() {
			t.Fatalf("expected partial parse up to target, len=%d parsed=%v", len(arr.value), arr.parsed.Load())
		}
	})

	t.Run("invalid element stores error", func(t *testing.T) {
		arr := NewArrayNode(nil, []byte(`[truX]`), nil).(*arrayNode)
		arr.lazyParseIndex(0)
		if arr.Error() == nil {
			t.Fatal("expected malformed element error")
		}
	})
}

func TestArrayLazyParsePathBranches(t *testing.T) {
	t.Run("no raw marks parsed", func(t *testing.T) {
		node := NewArrayNode(nil, nil, nil).(*arrayNode)
		node.lazyParsePath([]string{"0"})
		if !node.parsed.Load() {
			t.Fatal("expected parsed to be set when raw is empty")
		}
	})

	t.Run("empty path full parses", func(t *testing.T) {
		node, err := Parse([]byte(`[1,2]`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		arr := node.(*arrayNode)
		arr.lazyParsePath(nil)
		if !arr.parsed.Load() || len(arr.value) != 2 {
			t.Fatalf("expected full parse, parsed=%v len=%d", arr.parsed.Load(), len(arr.value))
		}
	})

	t.Run("invalid and negative indices fall back", func(t *testing.T) {
		for _, path := range [][]string{{"x"}, {"-1"}} {
			node, err := Parse([]byte(`[1,2]`))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			arr := node.(*arrayNode)
			arr.lazyParsePath(path)
			if !arr.parsed.Load() {
				t.Fatalf("expected fallback parse for path %v", path)
			}
		}
	})

	t.Run("already has enough elements", func(t *testing.T) {
		node, err := Parse([]byte(`[1,2,3]`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		arr := node.(*arrayNode)
		arr.value = []core.Node{NewNumberNode(arr, []byte("1"), nil), NewNumberNode(arr, []byte("2"), nil)}
		arr.lazyParsePath([]string{"1"})
		if len(arr.value) != 2 {
			t.Fatalf("expected no additional parsing, len=%d", len(arr.value))
		}
	})

	t.Run("malformed raw falls back", func(t *testing.T) {
		node := NewArrayNode(nil, []byte(`oops`), nil).(*arrayNode)
		node.lazyParsePath([]string{"0"})
		if !node.parsed.Load() {
			t.Fatal("expected malformed raw to trigger fallback parse")
		}
	})
}

func TestObjectLazyParsePathBranches(t *testing.T) {
	t.Run("error node returns immediately", func(t *testing.T) {
		node := &objectNode{baseNode: baseNode{err: sharedInvalidNode().Error()}}
		if got := node.GetWithPath("a", []string{"a"}); got != node {
			t.Fatal("expected errored GetWithPath to return self")
		}
		if got := node.LazyGet("a"); got != node {
			t.Fatal("expected errored LazyGet to return self")
		}
	})

	t.Run("no raw marks parsed", func(t *testing.T) {
		node := NewObjectNode(nil, nil, nil).(*objectNode)
		node.lazyParsePath([]string{"a"})
		if !node.parsed.Load() {
			t.Fatal("expected parsed to be set when raw is empty")
		}
	})

	t.Run("empty path full parses", func(t *testing.T) {
		node, err := Parse([]byte(`{"a":1}`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		obj := node.(*objectNode)
		obj.lazyParsePath(nil)
		if !obj.parsed.Load() || obj.Get("a").Int() != 1 {
			t.Fatalf("expected full parse, parsed=%v", obj.parsed.Load())
		}
	})

	t.Run("parsed object short-circuits", func(t *testing.T) {
		node, err := MustParse([]byte(`{"a":1}`))
		if err != nil {
			t.Fatalf("MustParse failed: %v", err)
		}
		obj := node.(*objectNode)
		obj.lazyParsePath([]string{"a"})
		if obj.Get("a").Int() != 1 {
			t.Fatal("expected parsed object get to keep working")
		}
	})

	t.Run("malformed raw falls back", func(t *testing.T) {
		node := NewObjectNode(nil, []byte(`oops`), nil).(*objectNode)
		obj := node
		obj.lazyParsePath([]string{"a"})
		if !obj.parsed.Load() {
			t.Fatal("expected malformed raw to trigger fallback parse")
		}
	})

	t.Run("missing key returns invalid", func(t *testing.T) {
		node, err := Parse([]byte(`{"a":1}`))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		obj := node.(*objectNode)
		if got := obj.GetWithPath("missing", []string{"missing"}); got.IsValid() {
			t.Fatal("expected missing GetWithPath result to be invalid")
		}
		if got := obj.LazyGet("missing"); got.IsValid() {
			t.Fatal("expected missing LazyGet result to be invalid")
		}
	})
}

func TestIteratorErrorAndEdgeBranches(t *testing.T) {
	objIter := (&objectIterator{rawMode: true, raw: []byte(`x`), pos: 0})
	if objIter.Next() {
		t.Fatal("expected malformed object iterator to stop")
	}
	if objIter.Err() == nil {
		t.Fatal("expected object iterator error")
	}
	if objIter.KeyRaw() != nil || objIter.ValueRaw() != nil {
		t.Fatal("expected nil raw accessors when iterator errored")
	}

	parsedObjIter := &objectIterator{rawMode: false, node: NewObjectNode(nil, nil, nil).(*objectNode), curKey: "missing"}
	if parsedObjIter.ParseValue().IsValid() {
		t.Fatal("expected invalid parsed object iterator ParseValue for missing key")
	}

	arrIter := (&arrayIterator{rawMode: true, raw: []byte(`x`), pos: 0})
	if arrIter.Next() {
		t.Fatal("expected malformed array iterator to stop")
	}
	if arrIter.Err() == nil {
		t.Fatal("expected array iterator error")
	}
	if arrIter.ValueRaw() != nil {
		t.Fatal("expected nil array raw value on iterator error")
	}

	parsedArrIter := &arrayIterator{rawMode: false, node: NewArrayNode(nil, nil, nil).(*arrayNode), curIndex: 3}
	if parsedArrIter.ParseValue().IsValid() {
		t.Fatal("expected invalid parsed array iterator ParseValue for out-of-range index")
	}
}
