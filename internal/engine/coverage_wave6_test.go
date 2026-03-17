package engine

import "testing"

func TestNumberRawFloatAndObjectHelpersMoreBranches(t *testing.T) {
	badNumber := NewNumberNode(nil, []byte("not-a-number"), nil).(*numberNode)
	if got, ok := badNumber.RawFloat(); ok || got != 0 {
		t.Fatalf("expected invalid RawFloat parse failure, got %v %v", got, ok)
	}

	if got := findMatchingQuote([]byte(`"unterminated`), 0); got != -1 {
		t.Fatalf("expected unterminated quote, got %d", got)
	}
	if got := findMatchingQuote([]byte(`x`), 0); got != -1 {
		t.Fatalf("expected non-quote start failure, got %d", got)
	}
	if got := findValueEnd([]byte(`123}`), 0); got != 2 {
		t.Fatalf("unexpected object delimiter value end: %d", got)
	}
	if got := findValueEnd([]byte(`123]`), 0); got != 2 {
		t.Fatalf("unexpected array delimiter value end: %d", got)
	}

	obj := NewObjectNode(nil, nil, nil).(*objectNode)
	obj.addChild("a", NewStringNode(nil, "x", nil))
	if len(obj.value) != 1 || obj.value["a"].Parent() != obj {
		t.Fatal("expected object addChild to initialize map and parent")
	}

	arr := NewArrayNode(nil, nil, nil).(*arrayNode)
	arr.value = nil
	arr.addChild(NewStringNode(nil, "x", nil))
	if len(arr.value) != 1 || arr.value[0].Parent() != arr {
		t.Fatal("expected array addChild to initialize slice and parent")
	}
}

func TestIteratorAndStringRemainingBranches(t *testing.T) {
	rawObj := &objectIterator{rawMode: true, raw: []byte(`{"a":truX}`), valStart: 5, valEnd: 8, node: NewObjectNode(nil, nil, nil).(*objectNode)}
	if got := rawObj.ParseValue(); got.IsValid() {
		t.Fatal("expected raw object iterator ParseValue failure")
	}

	rawArr := &arrayIterator{rawMode: true, raw: []byte(`[truX]`), valStart: 1, valEnd: 4, node: NewArrayNode(nil, nil, nil).(*arrayNode)}
	if got := rawArr.ParseValue(); got.IsValid() {
		t.Fatal("expected raw array iterator ParseValue failure")
	}

	objNode, err := MustParse([]byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	obj := objNode.(*objectNode)
	obj.value["a"].SetValue(2)
	if got := obj.String(); got != `{"a":2}` {
		t.Fatalf("expected dirty child object String rebuild, got %q", got)
	}

	emptyRaw := NewRawStringNode(nil, []byte(`""`), 1, 1, false, nil).(*stringNode)
	if got := emptyRaw.Raw(); got != "" {
		t.Fatalf("expected empty raw string Raw to be empty, got %q", got)
	}
}