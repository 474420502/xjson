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

	compiled, err := CompileQuery("/a/b[1]")
	if err != nil {
		t.Fatalf("CompileQuery failed: %v", err)
	}
	if compiled.fastPlan == nil || len(compiled.fastSteps) != 3 || compiled.specialized == nil || compiled.Path() != "/a/b[1]" {
		t.Fatalf("unexpected compiled query fast path state: %#v", compiled)
	}

	compiledGeneric, err := CompileQuery("/a/../a")
	if err != nil {
		t.Fatalf("CompileQuery generic path failed: %v", err)
	}
	if compiledGeneric.fastPlan != nil || len(compiledGeneric.tokens) == 0 {
		t.Fatalf("expected generic compiled query tokens, got %#v", compiledGeneric)
	}
}

func TestFlattenFastQueryPlanAndSteps(t *testing.T) {
	plan, ok := compileFastQueryPlan("/a/b[2]/name")
	if !ok {
		t.Fatal("expected fast query plan")
	}
	steps := flattenFastQueryPlan(plan)
	if len(steps) != 4 {
		t.Fatalf("unexpected step count: %d", len(steps))
	}
	if steps[0].kind != fastQueryStepKey || steps[0].key != "a" {
		t.Fatalf("unexpected first step: %#v", steps[0])
	}
	if steps[2].kind != fastQueryStepIndex || steps[2].index != 2 {
		t.Fatalf("unexpected index step: %#v", steps[2])
	}

	root, err := MustParse([]byte(`{"a":{"b":[{"name":"x"},{"name":"z"},{"name":"n"}]}}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	if got := executeFastQuerySteps(root, steps); !got.IsValid() || got.String() != "n" {
		t.Fatalf("unexpected fast step result: valid=%v val=%q err=%v", got.IsValid(), got.String(), got.Error())
	}

	specialized := buildSpecializedFastQuery(steps)
	if specialized == nil || specialized.kind != specializedFastQueryKeysIndexKeys {
		t.Fatalf("unexpected specialized query shape: %#v", specialized)
	}
	if len(specialized.preSuffixes) != 2 || specialized.arraySuffix != "[2]/name" || len(specialized.postSuffixes) != 1 {
		t.Fatalf("unexpected specialized suffix metadata: %#v", specialized)
	}
	if got := executeSpecializedFastQuery(root, specialized); !got.IsValid() || got.String() != "n" {
		t.Fatalf("unexpected specialized fast step result: valid=%v val=%q err=%v", got.IsValid(), got.String(), got.Error())
	}

	keyOnlyPlan, ok := compileFastQueryPlan("/a/x/y")
	if !ok {
		t.Fatal("expected key-only fast query plan")
	}
	keyOnlySpec := buildSpecializedFastQuery(flattenFastQueryPlan(keyOnlyPlan))
	if keyOnlySpec == nil || keyOnlySpec.kind != specializedFastQueryAllKeys {
		t.Fatalf("unexpected key-only specialized query shape: %#v", keyOnlySpec)
	}
	if len(keyOnlySpec.suffixes) != 3 || keyOnlySpec.suffixes[0] != "/a/x/y" || keyOnlySpec.suffixes[2] != "/y" {
		t.Fatalf("unexpected key-only suffixes: %#v", keyOnlySpec.suffixes)
	}
}

func TestCompiledQueryExecution(t *testing.T) {
	root, err := MustParse([]byte(`{"a":{"b":[{"name":"x"},{"name":"y"}]},"arr":[1,2,3]}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}

	fast, err := CompileQuery("/a/b[1]/name")
	if err != nil {
		t.Fatalf("CompileQuery fast path failed: %v", err)
	}
	if got := fast.Query(root); !got.IsValid() || got.String() != "y" {
		t.Fatalf("unexpected fast compiled query result: valid=%v val=%q err=%v", got.IsValid(), got.String(), got.Error())
	}

	generic, err := CompileQuery("/a/b[1]/../[0]/name")
	if err != nil {
		t.Fatalf("CompileQuery generic failed: %v", err)
	}
	if got := generic.Query(root); !got.IsValid() || got.String() != "x" {
		t.Fatalf("unexpected generic compiled query result: valid=%v val=%q err=%v", got.IsValid(), got.String(), got.Error())
	}

	if got := (*CompiledQuery)(nil).Query(root); got.IsValid() {
		t.Fatal("expected nil compiled query to return invalid node")
	}

	ResetQueryCache(root)
	if got := fast.Query(root); !got.IsValid() || got.String() != "y" {
		t.Fatalf("unexpected repeated fast compiled query result after root reset: valid=%v val=%q err=%v", got.IsValid(), got.String(), got.Error())
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
