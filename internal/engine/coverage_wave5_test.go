package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestApplySimpleQueryMoreBranches(t *testing.T) {
	funcs := &map[string]core.UnaryPathFunc{
		"self": func(n core.Node) core.Node {
			return n
		},
	}
	root, err := ParseWithFuncs([]byte(`{"store":{"books":[{"title":"a","price":10},{"title":"b","price":20}],"meta":{"name":"shop"}},"tree":{"id":1,"child":{"id":2}}}`), funcs)
	if err != nil {
		t.Fatalf("ParseWithFuncs failed: %v", err)
	}

	if got := applySimpleQuery(root, "/store/books/title"); !got.IsValid() || got.Len() != 2 {
		t.Fatalf("expected array key projection result, got len=%d err=%v", got.Len(), got.Error())
	}
	if got := applySimpleQuery(root, "/store/books[0:1]"); !got.IsValid() || got.Len() != 1 {
		t.Fatalf("expected slice result, got len=%d err=%v", got.Len(), got.Error())
	}
	if got := applySimpleQuery(root, "/store/*"); !got.IsValid() || got.Len() != 2 {
		t.Fatalf("expected wildcard result, got len=%d err=%v", got.Len(), got.Error())
	}
	if got := applySimpleQuery(root, "/store/meta[@self]/name"); !got.IsValid() || got.String() != "shop" {
		t.Fatalf("expected function path result, got %q err=%v", got.String(), got.Error())
	}
	if got := applySimpleQuery(root, "/tree//id"); !got.IsValid() || got.Len() != 2 {
		t.Fatalf("expected recursive query result, got len=%d err=%v", got.Len(), got.Error())
	}
	if got := applySimpleQuery(root, "/tree/child/.."); !got.IsValid() || got.Get("id").Int() != 1 {
		t.Fatalf("expected parent query result, got %v err=%v", got.Interface(), got.Error())
	}
	if got := applySimpleQuery(root, "/.."); got.IsValid() {
		t.Fatal("expected parent-above-root query to be invalid")
	}
	if got := applySimpleQuery(root.Query("/store/meta/name"), "/[0:1]"); got.IsValid() {
		t.Fatal("expected slice on non-array to be invalid")
	}
}

func TestObjectStringMapAndChildHelpers(t *testing.T) {
	root, err := MustParse([]byte(`{"obj":{"a":1},"arr":[1]}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	obj := root.Get("obj").(*objectNode)
	if got := obj.AsMap(); len(got) != 1 || got["a"].Int() != 1 {
		t.Fatalf("unexpected AsMap result: %#v", got)
	}
	if got := obj.MustAsMap(); len(got) != 1 || got["a"].Int() != 1 {
		t.Fatalf("unexpected MustAsMap result: %#v", got)
	}
	if raw := obj.String(); raw != `{"a":1}` {
		t.Fatalf("expected pristine object String to use raw bytes, got %q", raw)
	}

	obj.addChild("b", NewNumberNode(nil, []byte("2"), nil))
	if obj.value["b"].Parent() != obj {
		t.Fatal("expected object addChild to reparent child")
	}
	obj.isDirty = true
	if raw := obj.String(); raw != `{"a":1,"b":2}` && raw != `{"b":2,"a":1}` {
		t.Fatalf("expected dirty object String to rebuild content, got %q", raw)
	}

	arr := root.Get("arr").(*arrayNode)
	arr.addChild(NewNumberNode(nil, []byte("2"), nil))
	if arr.value[1].Parent() != arr {
		t.Fatal("expected array addChild to reparent child")
	}

	errObj := &objectNode{baseNode: baseNode{err: sharedInvalidNode().Error()}}
	if errObj.AsMap() != nil {
		t.Fatal("expected errored object AsMap to return nil")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected errored object MustAsMap to panic")
		}
	}()
	_ = errObj.MustAsMap()
}

func TestBaseQueryCallFuncAndIteratorEdges(t *testing.T) {
	errBase := &baseNode{err: sharedInvalidNode().Error()}
	if got := errBase.Query("/a"); got.IsValid() {
		t.Fatal("expected errored Query to return invalid")
	}

	funcs := map[string]core.UnaryPathFunc{"id": func(n core.Node) core.Node { return n }}
	withMissingSelf := &baseNode{funcs: &funcs}
	if got := withMissingSelf.CallFunc("id"); got.IsValid() {
		t.Fatal("expected CallFunc with nil self to be invalid")
	}

	parsedObjIter := &objectIterator{rawMode: false, node: NewObjectNode(nil, nil, nil).(*objectNode), curKey: "missing"}
	if parsedObjIter.ValueRaw() != nil {
		t.Fatal("expected parsed object iterator missing ValueRaw to be nil")
	}
	parsedArrIter := &arrayIterator{rawMode: false, node: NewArrayNode(nil, nil, nil).(*arrayNode), curIndex: -1}
	if parsedArrIter.ValueRaw() != nil {
		t.Fatal("expected parsed array iterator invalid index ValueRaw to be nil")
	}
	if got := (&arrayIterator{rawMode: false}).ParseValue(); got.IsValid() {
		t.Fatal("expected parsed array iterator nil node ParseValue to be invalid")
	}
}

func TestArrayAppendAndMustArrayMoreBranches(t *testing.T) {
	root, err := MustParse([]byte(`{"arr":[1]}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	arr := root.Get("arr").(*arrayNode)
	if got := arr.Append(2); !got.IsValid() || arr.Len() != 2 || arr.Index(1).Int() != 2 {
		t.Fatalf("expected Append to succeed, len=%d err=%v", arr.Len(), got.Error())
	}
	if !arr.isDirty || !root.(*objectNode).isDirty {
		t.Fatal("expected Append to mark array and ancestors dirty")
	}

	errArr := &arrayNode{baseNode: baseNode{err: sharedInvalidNode().Error()}}
	if got := errArr.Append(1); got != errArr {
		t.Fatal("expected errored Append to return self")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected MustArray to panic on errored array")
		}
	}()
	_ = errArr.MustArray()
}