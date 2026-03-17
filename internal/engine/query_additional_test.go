package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestRawAndFastQueryHelpers(t *testing.T) {
	root, err := Parse([]byte(`{
		"store": {
			"books": [
				{"title": "Book 1", "price": 10},
				{"title": "Book 2", "price": 20}
			],
			"meta": {"name": "shop"}
		}
	}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if got := tryRawDirectPath(root, "/store/books[1]/title"); !got.IsValid() || got.String() != "Book 2" {
		t.Fatalf("unexpected tryRawDirectPath result: valid=%v val=%q err=%v", got.IsValid(), got.String(), got.Error())
	}
	if got := tryRawDirectPath(root, "/store/books[2]/title"); got.IsValid() {
		t.Fatal("expected invalid result for missing array element")
	}
	if got := tryRawDirectPath(root, "/store/*/title"); got != nil {
		t.Fatalf("expected nil for unsupported raw direct path, got %v", got)
	}

	if got := tryFastSlashQuery(root, "/store/meta/name"); !got.IsValid() || got.String() != "shop" {
		t.Fatalf("unexpected tryFastSlashQuery result: %q err=%v", got.String(), got.Error())
	}
	if got := tryFastSlashQuery(root, "/store/books[0]"); got != nil {
		t.Fatal("expected nil for bracket path in slash fast path")
	}
	if got := tryFastSlashQuery(root, "/store/missing/name"); got == nil || got.IsValid() {
		t.Fatal("expected invalid result for missing intermediate slash path")
	}
	if got := tryFastSlashQuery(root, ""); got != nil {
		t.Fatal("expected empty slash fast path to be rejected")
	}
}

func TestQueryPlanAndCompiledQueryCaches(t *testing.T) {
	plan, ok := compileFastQueryPlan("/a/b[10]/c[-1]")
	if !ok || len(plan.segments) != 3 {
		t.Fatalf("unexpected plan: ok=%v len=%d", ok, len(plan.segments))
	}
	if plan.segments[1].key != "b" || len(plan.segments[1].indices) != 1 || plan.segments[1].indices[0] != 10 {
		t.Fatalf("unexpected middle segment: %#v", plan.segments[1])
	}
	if plan.segments[2].key != "c" || len(plan.segments[2].indices) != 1 || plan.segments[2].indices[0] != -1 {
		t.Fatalf("unexpected final segment: %#v", plan.segments[2])
	}

	badPaths := []string{"/a/*", "/a[@x]", "/a//b", "../a", "/a["}
	for _, path := range badPaths {
		if _, ok := compileFastQueryPlan(path); ok {
			t.Fatalf("expected compileFastQueryPlan to reject %q", path)
		}
	}

	if _, ok := getFastQueryPlan("/a/b[1]"); !ok {
		t.Fatal("expected getFastQueryPlan cache fill")
	}
	tokens, err := ParseQuery("/a/b[1]")
	if err != nil || len(tokens) == 0 {
		t.Fatalf("ParseQuery failed: %v", err)
	}
	if cached, ok := getCachedCompiledQuery("/a/b[1]"); !ok || len(cached) != len(tokens) {
		t.Fatalf("expected compiled query cache hit, got ok=%v len=%d", ok, len(cached))
	}
}

func TestQueryLowLevelHelpers(t *testing.T) {
	obj := NewObjectNode(nil, []byte(`{"a":1,"b":"x"}`), nil).(*objectNode)
	obj.mu.Lock()
	child, found, ok := fastScanObjectChildLocked(obj, "b")
	obj.mu.Unlock()
	if !ok || !found || child.String() != "x" {
		t.Fatalf("unexpected fastScanObjectChildLocked result: ok=%v found=%v val=%q", ok, found, child.String())
	}

	if match, decoded, err := matchObjectKey("a\"b", []byte(`a\"b`)); err != nil || !match || decoded != `a"b` {
		t.Fatalf("unexpected matchObjectKey result: match=%v decoded=%q err=%v", match, decoded, err)
	}
	if _, _, err := matchObjectKey("x", []byte(`\uZZZZ`)); err == nil {
		t.Fatal("expected matchObjectKey error")
	}

	if node := directObjectChild(obj, "missing"); node.IsValid() {
		t.Fatal("expected invalid directObjectChild result")
	}

	arrRoot, err := Parse([]byte(`[10,20]`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	arr := arrRoot.(*arrayNode)
	arr.lazyParse()
	if node := directArrayChild(arr, -1); !node.IsValid() || node.Int() != 20 {
		t.Fatalf("unexpected directArrayChild result: %v", node.Int())
	}

	constructed := fastConstructObjectChild(obj, []byte(`true`))
	if !constructed.IsValid() || constructed.Type() != core.Bool {
		t.Fatalf("unexpected constructed node: %v", constructed.Type())
	}
	if constructed := fastConstructObjectChild(obj, []byte(`"y"`)); !constructed.IsValid() || constructed.String() != "y" {
		t.Fatalf("unexpected string constructed node: valid=%v val=%q err=%v", constructed.IsValid(), constructed.String(), constructed.Error())
	}
	if constructed := fastConstructObjectChild(obj, []byte(`null`)); !constructed.IsValid() || constructed.Type() != core.Null {
		t.Fatalf("unexpected null constructed node: %v", constructed.Type())
	}
	if constructed := fastConstructObjectChild(obj, []byte(`{"k":1}`)); !constructed.IsValid() || constructed.Type() != core.Object {
		t.Fatalf("unexpected object constructed node: %v", constructed.Type())
	}
	if constructed := fastConstructObjectChild(obj, []byte(`[1]`)); !constructed.IsValid() || constructed.Type() != core.Array {
		t.Fatalf("unexpected array constructed node: %v", constructed.Type())
	}

	results := recursiveSearch(obj, "")
	if !results.IsValid() || results.Type() != core.Array {
		t.Fatalf("unexpected recursiveSearch result: valid=%v type=%v", results.IsValid(), results.Type())
	}

	if raw, pos, ok := initObjectRawScanLocked(NewObjectNode(nil, []byte(`{"x":1}`), nil).(*objectNode)); !ok || len(raw) == 0 || pos == 0 {
		t.Fatalf("unexpected initObjectRawScanLocked result: ok=%v len=%d pos=%d", ok, len(raw), pos)
	}
	if compareStringBytes("", nil) != true || compareStringBytes("a", []byte("b")) {
		t.Fatal("unexpected compareStringBytes results")
	}
	if got := findMatchingQuote([]byte(`"a\"b"`), 0); got != 5 {
		t.Fatalf("unexpected findMatchingQuote result: %d", got)
	}
	if got := findValueEnd([]byte(`123,456`), 0); got != 2 {
		t.Fatalf("unexpected findValueEnd result: %d", got)
	}
}
