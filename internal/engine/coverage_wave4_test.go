package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestObjectGetMoreBranches(t *testing.T) {
	errNode := &objectNode{baseNode: baseNode{err: sharedInvalidNode().Error()}}
	if got := errNode.Get("a"); got != errNode {
		t.Fatal("expected errored object Get to return itself")
	}

	parsed, err := MustParse([]byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	if got := parsed.(*objectNode).Get("a"); !got.IsValid() || got.Int() != 1 {
		t.Fatalf("expected parsed object Get hit, got %v err=%v", got.Interface(), got.Error())
	}
	if got := parsed.(*objectNode).Get("missing"); got.IsValid() {
		t.Fatal("expected parsed object Get miss to be invalid")
	}

	constructed := NewObjectNode(nil, nil, nil).(*objectNode)
	constructed.value = map[string]core.Node{"a": NewNumberNode(constructed, []byte("1"), nil)}
	if got := constructed.Get("a"); !got.IsValid() || got.Int() != 1 {
		t.Fatalf("expected constructed object Get hit, got %v err=%v", got.Interface(), got.Error())
	}

	rawDone := NewObjectNode(nil, []byte(`{"a":1}`), nil).(*objectNode)
	rawDone.rawDone = true
	if got := rawDone.Get("missing"); got.IsValid() {
		t.Fatal("expected rawDone object Get miss to be invalid")
	}

	malformed := NewObjectNode(nil, []byte(`{"a"`), nil).(*objectNode)
	if got := malformed.Get("a"); got.IsValid() {
		t.Fatal("expected malformed object Get to be invalid")
	}
}

func TestBaseNodeAccessorBranches(t *testing.T) {
	errBase := &baseNode{err: sharedInvalidNode().Error()}
	if got := errBase.Parent(); got == nil || got.IsValid() {
		t.Fatal("expected errored Parent to return invalid self")
	}

	rawNode := &baseNode{raw: []byte("abcd"), start: -1, end: 99}
	if got := string(rawNode.RawBytes()); got != "abcd" {
		t.Fatalf("unexpected clamped RawBytes: %q", got)
	}
	if got := (&baseNode{raw: []byte("abcd"), start: 3, end: 1}).RawBytes(); got != nil {
		t.Fatal("expected inverted RawBytes bounds to return nil")
	}

	arr := NewArrayNode(nil, nil, nil).(*arrayNode)
	if got := arr.MustArray(); got == nil || len(got) != 0 {
		t.Fatalf("expected empty constructed MustArray, got %#v", got)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected invalid string MustString to panic")
		}
	}()
	_ = (&stringNode{baseNode: baseNode{err: sharedInvalidNode().Error()}}).MustString()
}

func TestQueryParserCacheBranches(t *testing.T) {
	oldCompiled := compiledQueryCache.m
	compiledQueryCache.m = make(map[string][]queryToken)
	defer func() { compiledQueryCache.m = oldCompiled }()

	cacheCompiledQuery("/a", []queryToken{{Op: OpKey, Value: "a"}})
	cacheCompiledQuery("/a", []queryToken{{Op: OpKey, Value: "b"}})
	if got, ok := getCachedCompiledQuery("/a"); !ok || len(got) != 1 || got[0].Value.(string) != "a" {
		t.Fatalf("expected cacheCompiledQuery to preserve first insert, got %#v ok=%v", got, ok)
	}

	for i := 0; i < maxCompiledQueryEntries; i++ {
		compiledQueryCache.m[string(rune(i+1))] = []queryToken{{Op: OpKey, Value: i}}
	}
	cacheCompiledQuery("overflow", []queryToken{{Op: OpKey, Value: "x"}})
	if _, ok := compiledQueryCache.m["overflow"]; ok {
		t.Fatal("expected cacheCompiledQuery to refuse inserts when full")
	}
}

func TestSIMDQuoteWordMoreBranches(t *testing.T) {
	unterminated := append([]byte{byte('"')}, make([]byte, 80)...)
	for i := 1; i < len(unterminated); i++ {
		unterminated[i] = 'a'
	}
	if got := findQuoteWord(unterminated, 0); got != -1 {
		t.Fatalf("expected unterminated quote word scan to fail, got %d", got)
	}

	escaped := append([]byte{byte('"')}, make([]byte, 90)...)
	for i := 1; i < len(escaped); i++ {
		escaped[i] = 'a'
	}
	escaped[8] = '\\'
	escaped[9] = '"'
	escaped[len(escaped)-1] = '"'
	if got := findQuoteWord(escaped, 0); got != len(escaped)-1 {
		t.Fatalf("expected escaped quote scan to find terminal quote, got %d", got)
	}
}