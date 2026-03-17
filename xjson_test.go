package xjson

import "testing"

func TestParseAndMustParseWrappers(t *testing.T) {
	root, err := Parse(`{"a":1}`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if root.Type() != Object || root.Query("/a").Int() != 1 {
		t.Fatalf("unexpected Parse result: type=%v value=%d", root.Type(), root.Query("/a").Int())
	}

	root, err = Parse([]byte(`{"a":2}`))
	if err != nil || root.Query("/a").Int() != 2 {
		t.Fatalf("Parse []byte failed: %v", err)
	}

	must, err := MustParse(`{"a":{"b":3}}`)
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	if must.Query("/a/b").Int() != 3 {
		t.Fatalf("unexpected MustParse result: %d", must.Query("/a/b").Int())
	}

	if _, err := Parse(123); err == nil {
		t.Fatal("expected Parse type error")
	}
	if _, err := MustParse(123); err == nil {
		t.Fatal("expected MustParse type error")
	}
	if _, err := Parse(""); err == nil {
		t.Fatal("expected Parse empty data error")
	}
	if _, err := MustParse([]byte{}); err == nil {
		t.Fatal("expected MustParse empty data error")
	}
}

func TestNodeWrapperSetByPath(t *testing.T) {
	root, err := MustParse(`{"cfg":{}}`)
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}
	updated := root.(nodeWrapper).SetByPath("/cfg/name", "demo")
	if !updated.IsValid() {
		t.Fatalf("SetByPath failed: %v", updated.Error())
	}
	if got := root.Query("/cfg/name").String(); got != "demo" {
		t.Fatalf("expected demo, got %q", got)
	}
}

func TestPreparedQueryWrapper(t *testing.T) {
	root, err := MustParse(`{"store":{"books":[{"title":"Book 1"},{"title":"Book 2"}]}}`)
	if err != nil {
		t.Fatalf("MustParse failed: %v", err)
	}

	pq, err := CompileQuery("/store/books[1]/title")
	if err != nil {
		t.Fatalf("CompileQuery failed: %v", err)
	}
	if pq.Path() != "/store/books[1]/title" {
		t.Fatalf("unexpected prepared path: %q", pq.Path())
	}
	if got := pq.Query(root); !got.IsValid() || got.String() != "Book 2" {
		t.Fatalf("unexpected prepared query result: valid=%v val=%q err=%v", got.IsValid(), got.String(), got.Error())
	}

	if _, err := CompileQuery("/store["); err == nil {
		t.Fatal("expected CompileQuery syntax error")
	}

	must := MustCompileQuery("/store/books[0]/title")
	if got := must.Query(root).String(); got != "Book 1" {
		t.Fatalf("unexpected MustCompileQuery result: %q", got)
	}
}
