package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestBaseNodeRawStringAndFloatSwitches(t *testing.T) {
	number := NewNumberNode(nil, []byte("12.5"), nil).(*numberNode)
	if got, ok := (&baseNode{self: number}).RawFloat(); !ok || got != 12.5 {
		t.Fatalf("unexpected number RawFloat: %v %v", got, ok)
	}

	boolean := NewBoolNode(nil, true, nil).(*boolNode)
	if _, ok := (&baseNode{self: boolean}).RawFloat(); ok {
		t.Fatal("expected bool RawFloat to be unavailable")
	}
	if got, ok := (&baseNode{self: boolean}).RawString(); !ok || got != "true" {
		t.Fatalf("unexpected bool RawString: %q %v", got, ok)
	}

	stringNode := NewStringNode(nil, "hello", nil).(*stringNode)
	if _, ok := (&baseNode{self: stringNode}).RawFloat(); ok {
		t.Fatal("expected string RawFloat to be unavailable")
	}
	if got, ok := (&baseNode{self: stringNode}).RawString(); !ok || got != "hello" {
		t.Fatalf("unexpected string RawString: %q %v", got, ok)
	}

	nullNode := NewNullNode(nil, nil).(*nullNode)
	if got, ok := (&baseNode{self: nullNode}).RawString(); !ok || got != "null" {
		t.Fatalf("unexpected null RawString: %q %v", got, ok)
	}
	if _, ok := (&baseNode{self: nullNode}).RawFloat(); ok {
		t.Fatal("expected null RawFloat to be unavailable")
	}
	if got, ok := (&baseNode{raw: []byte("raw")}).RawString(); !ok || got != "raw" {
		t.Fatalf("unexpected fallback RawString: %q %v", got, ok)
	}
}

func TestSimpleNodeTimeAndStringBranches(t *testing.T) {
	empty := NewRawStringNode(nil, []byte(`""`), 1, 1, false, nil)
	if got := empty.String(); got != "" {
		t.Fatalf("expected empty raw string to decode to empty string, got %q", got)
	}

	badTime := NewStringNode(nil, "not-a-time", nil)
	if tm := badTime.Time(); !tm.IsZero() {
		t.Fatal("expected zero time for invalid timestamp")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected MustTime to panic on invalid timestamp")
		}
	}()
	_ = badTime.MustTime()
}

func TestObjectAndArrayMutationBranches(t *testing.T) {
	root, err := MustParse([]byte(`{"obj":{"n":1},"arr":[1]}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}

	obj := root.Query("/obj").(*objectNode)
	if got := obj.Set("n", 2); !got.IsValid() || root.Query("/obj/n").Int() != 2 {
		t.Fatalf("expected scalar object mutation to succeed: %v", got.Error())
	}
	if got := obj.Set("bad", struct{}{}); got.IsValid() {
		t.Fatal("expected invalid object child creation")
	}

	arr := root.Query("/arr").(*arrayNode)
	if got := arr.Append(struct{}{}); got.IsValid() {
		t.Fatal("expected invalid array append child")
	}
	if got := arr.Set("bad", 1); got.IsValid() {
		t.Fatal("expected invalid array set index")
	}
	if got := arr.Set("5", 1); got.IsValid() {
		t.Fatal("expected out of bounds array set")
	}
}

func TestFastPathVariantsAndRecursiveSearch(t *testing.T) {
	root, err := Parse([]byte(`{"a":{"b":{"c":1}},"items":[{"name":"x"}],"mixed":[{"k":1},[2]],"flag":true}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if got := tryFastSlashQuery(root, "/"); got != root {
		t.Fatal("expected root for slash-only path")
	}
	if got := tryFastSlashQuery(root, "/a/b/c"); !got.IsValid() || got.Int() != 1 {
		t.Fatalf("unexpected fast slash nested result: %v", got.Interface())
	}
	if got := tryFastSlashQuery(root.Query("/flag"), "/x"); got != nil {
		t.Fatal("expected nil for non-object slash fast path")
	}

	if got := tryRawDirectPath(root, "/items[0]/name"); !got.IsValid() || got.String() != "x" {
		t.Fatalf("unexpected raw direct path result: %q err=%v", got.String(), got.Error())
	}
	if got := tryRawDirectPath(root.Query("/flag"), "/x"); got != nil {
		t.Fatal("expected nil raw direct path for primitive start")
	}
	if got := tryRawDirectPath(root, "/items[-1]/name"); got.IsValid() {
		t.Fatal("expected invalid raw direct path for negative index")
	}

	parsed, err := MustParse([]byte(`{"top":{"k":1},"arr":[{"k":2}]}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	res := recursiveSearch(parsed, "k")
	if !res.IsValid() || res.Len() != 2 {
		t.Fatalf("unexpected recursiveSearch parsed result: len=%d err=%v", res.Len(), res.Error())
	}
}

func TestGetWithPathAndIteratorBranches(t *testing.T) {
	root, err := Parse([]byte(`{"obj":{"x":1},"arr":[1,2]}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	obj := root.(*objectNode)
	if got := obj.GetWithPath("obj", nil); !got.IsValid() {
		t.Fatal("expected GetWithPath with nil path to work")
	}
	if got := obj.Get("missing"); got.IsValid() {
		t.Fatal("expected missing object child to be invalid")
	}

	objIter := obj.Iter()
	for objIter.Next() {
		_ = objIter.KeyRaw()
		_ = objIter.ValueRaw()
		_ = objIter.ParseValue()
	}
	if objIter.Err() != nil {
		t.Fatalf("unexpected object iterator error: %v", objIter.Err())
	}

	arr := root.Query("/arr").(*arrayNode)
	arrIter := arr.Iter()
	for arrIter.Next() {
		_ = arrIter.Index()
		_ = arrIter.ValueRaw()
		_ = arrIter.ParseValue()
	}
	if arrIter.Err() != nil {
		t.Fatalf("unexpected array iterator error: %v", arrIter.Err())
	}

	nilObjIter := (*objectNode)(nil).Iter()
	if nilObjIter.Err() == nil {
		t.Fatal("expected nil object iterator error")
	}
	nilArrIter := (*arrayNode)(nil).Iter()
	if nilArrIter.Err() == nil {
		t.Fatal("expected nil array iterator error")
	}
}

func TestEngineTopLevelHelpers(t *testing.T) {
	funcs := &map[string]core.UnaryPathFunc{}
	if node, err := MustParseWithFuncs([]byte(`{"a":1}`), funcs); err != nil || node.Type() != core.Object {
		t.Fatalf("MustParseWithFuncs failed: %v", err)
	}
	if node, err := ParseWithFuncs([]byte(`[1]`), funcs); err != nil || node.Type() != core.Array {
		t.Fatalf("ParseWithFuncs failed: %v", err)
	}
	if getFirstNonWhitespaceChar([]byte(" \n\t")) != 0 {
		t.Fatal("expected zero char for whitespace-only input")
	}
}
