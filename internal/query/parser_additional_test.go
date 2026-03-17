package query

import "testing"

func TestParserHelperFunctions(t *testing.T) {
	p := NewParser("/a")
	if errs := p.Errors(); errs == nil || len(errs) != 0 {
		t.Fatalf("expected empty errors slice, got %#v", errs)
	}

	if segment, next, err := parseIdentifierSegment("abc/def", 0); err != nil || segment != "abc" || next != 3 {
		t.Fatalf("unexpected parseIdentifierSegment result: %q %d %v", segment, next, err)
	}
	if segment, next, err := parseIdentifierSegment("abc", 0); err != nil || segment != "abc" || next != 3 {
		t.Fatalf("unexpected parseIdentifierSegment eof result: %q %d %v", segment, next, err)
	}

	if key, next, err := parseQuotedKey(`'a\'b'`, 0); err != nil || key != "a'b" || next != len(`'a\'b'`) {
		t.Fatalf("unexpected parseQuotedKey single-quote result: %q %d %v", key, next, err)
	}
	if key, next, err := parseQuotedKey(`"a\"b"`, 0); err != nil || key != `a"b` || next != len(`"a\"b"`) {
		t.Fatalf("unexpected parseQuotedKey double-quote result: %q %d %v", key, next, err)
	}
	if _, _, err := parseQuotedKey(`'abc\`, 0); err == nil {
		t.Fatal("expected parseQuotedKey escape error")
	}

	if token, next, err := parseBracketExpression("[*]", 0); err != nil || token.Type != OpWildcard || next != 3 {
		t.Fatalf("unexpected wildcard bracket result: %#v %d %v", token, next, err)
	}
	if _, _, err := parseBracketExpression("[*x]", 0); err == nil {
		t.Fatal("expected wildcard bracket error")
	}
	if _, _, err := parseBracketExpression("[@name", 0); err == nil {
		t.Fatal("expected function closing bracket error")
	}
	if _, _, err := parseBracketExpression("['name'", 0); err == nil {
		t.Fatal("expected quoted key closing bracket error")
	}
	if _, _, err := parseBracketExpression("[a:b]", 0); err == nil {
		t.Fatal("expected invalid slice start error")
	}
	if _, _, err := parseBracketExpression("[1:b]", 0); err == nil {
		t.Fatal("expected invalid slice end error")
	}

	if _, ok := tryParseInt("-10"); !ok {
		t.Fatal("expected negative int parse to succeed")
	}
	if _, ok := tryParseInt(""); ok {
		t.Fatal("expected empty int parse to fail")
	}
	if _, ok := tryParseInt("1a"); ok {
		t.Fatal("expected invalid int parse to fail")
	}

	if !isIdentifier("abc_1") || isIdentifier("1abc") || isIdentifier("a-b") {
		t.Fatal("unexpected identifier classification")
	}
}