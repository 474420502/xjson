package engine

import (
	"bytes"
	"strings"
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestParserParseAndParseFull(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		p := newParser([]byte("   \n\t"), nil)
		if _, err := p.Parse(); err == nil {
			t.Fatal("expected Parse to fail on empty input")
		}
		p = newParser([]byte("   \n\t"), nil)
		if _, err := p.ParseFull(); err == nil {
			t.Fatal("expected ParseFull to fail on empty input")
		}
	})

	t.Run("parse primitive root", func(t *testing.T) {
		p := newParser([]byte(" 123 "), nil)
		node, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if node.Type() != core.Number || node.Int() != 123 {
			t.Fatalf("unexpected node: type=%v int=%d", node.Type(), node.Int())
		}

		p = newParser([]byte(` {"a":[1,true,null,"x"]} `), nil)
		node, err = p.ParseFull()
		if err != nil {
			t.Fatalf("ParseFull failed: %v", err)
		}
		if node.Type() != core.Object {
			t.Fatalf("expected object, got %v", node.Type())
		}
		if got := node.Query("/a[3]").String(); got != "x" {
			t.Fatalf("expected x, got %q", got)
		}
	})

	t.Run("invalid root character", func(t *testing.T) {
		p := newParser([]byte("?"), nil)
		if _, err := p.Parse(); err == nil || !strings.Contains(err.Error(), "invalid character") {
			t.Fatalf("expected invalid character error, got %v", err)
		}
	})
}

func TestParserValueHelpersAndErrors(t *testing.T) {
	t.Run("parse value end of input", func(t *testing.T) {
		p := newParser(nil, nil)
		if node := p.parseValue(nil); node.Error() == nil {
			t.Fatal("expected parseValue error")
		}
		if node := p.parseValueFull(nil); node.Error() == nil {
			t.Fatal("expected parseValueFull error")
		}
	})

	t.Run("object and array parse errors", func(t *testing.T) {
		cases := []struct {
			name string
			data string
			fn   func(*parser) core.Node
			err  string
		}{
			{"unterminated object lazy", `{`, func(p *parser) core.Node { return p.parseObject(nil) }, "unterminated object"},
			{"missing colon", `{"a" 1}`, func(p *parser) core.Node { return p.parseObjectFull(nil) }, "missing ':'"},
			{"missing comma object", `{"a":1 "b":2}`, func(p *parser) core.Node { return p.parseObjectFull(nil) }, "missing ','"},
			{"missing comma array", `[1 2]`, func(p *parser) core.Node { return p.parseArrayFull(nil) }, "missing ','"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				p := newParser([]byte(tc.data), nil)
				node := tc.fn(p)
				if node.Error() == nil || !strings.Contains(node.Error().Error(), tc.err) {
					t.Fatalf("expected error containing %q, got %v", tc.err, node.Error())
				}
			})
		}
	})

	t.Run("parseArrayFull success and parsed flag", func(t *testing.T) {
		p := newParser([]byte(`[1,{"a":2}]`), nil)
		node := p.parseArrayFull(nil)
		if !node.IsValid() || node.Type() != core.Array {
			t.Fatalf("expected valid array node, got type=%v err=%v", node.Type(), node.Error())
		}
		arr := node.(*arrayNode)
		if !arr.parsed.Load() || arr.Len() != 2 || arr.Index(1).Get("a").Int() != 2 {
			t.Fatalf("unexpected parseArrayFull result: parsed=%v len=%d", arr.parsed.Load(), arr.Len())
		}
	})

	t.Run("invalid bool and null", func(t *testing.T) {
		p := newParser([]byte("truX"), nil)
		if node := p.parseBool(nil); node.Error() == nil {
			t.Fatal("expected invalid boolean error")
		}
		p = newParser([]byte("nulX"), nil)
		if node := p.parseNull(nil); node.Error() == nil {
			t.Fatal("expected invalid null error")
		}
	})

	t.Run("unterminated string", func(t *testing.T) {
		p := newParser([]byte(`"abc`), nil)
		if node := p.parseString(nil); node.Error() == nil {
			t.Fatal("expected unterminated string")
		}
	})
}

func TestParserStringAndUnescapeHelpers(t *testing.T) {
	t.Run("parse escaped string", func(t *testing.T) {
		p := newParser([]byte(`"hello\nworld"`), nil)
		node := p.parseString(nil)
		if !node.IsValid() {
			t.Fatalf("parseString failed: %v", node.Error())
		}
		if got := node.String(); got != "hello\nworld" {
			t.Fatalf("unexpected string: %q", got)
		}
	})

	t.Run("unescape variants", func(t *testing.T) {
		cases := []struct {
			name string
			in   []byte
			want string
		}{
			{"plain", []byte("plain"), "plain"},
			{"newline", []byte(`a\nb`), "a\nb"},
			{"tab", []byte(`a\tb`), "a\tb"},
			{"quote", []byte(`a\"b`), `a"b`},
			{"slash", []byte(`a\/b`), `a/b`},
			{"unicode", []byte(`\u4f60`), "你"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				out, err := unescape(tc.in)
				if err != nil {
					t.Fatalf("unescape failed: %v", err)
				}
				if string(out) != tc.want {
					t.Fatalf("expected %q, got %q", tc.want, string(out))
				}

				var buf []byte
				out, err = unescapeWithBuffer(tc.in, &buf)
				if err != nil {
					t.Fatalf("unescapeWithBuffer failed: %v", err)
				}
				if string(out) != tc.want {
					t.Fatalf("expected %q, got %q", tc.want, string(out))
				}
			})
		}
	})

	t.Run("unescape error branches", func(t *testing.T) {
		badInputs := [][]byte{
			[]byte(`abc\`),
			[]byte(`\x`),
			[]byte(`\u12`),
			[]byte(`\uZZZZ`),
		}
		for _, input := range badInputs {
			if _, err := unescape(input); err == nil {
				t.Fatalf("expected unescape error for %q", input)
			}
			var buf []byte
			if _, err := unescapeWithBuffer(input, &buf); err == nil {
				t.Fatalf("expected unescapeWithBuffer error for %q", input)
			}
		}

		p := newParser([]byte(`[1,2`), nil)
		if node := p.parseArray(nil); node.Error() == nil || !strings.Contains(node.Error().Error(), "unterminated array") {
			t.Fatalf("expected unterminated array error, got %v", node.Error())
		}
	})
}

func TestParserCountHelpers(t *testing.T) {
	objectCases := []struct {
		data  string
		start int
		want  int
	}{
		{`{}`, 0, 0},
		{`{"a":1}`, 0, 0},
		{`{"a":1,"b":{"c":2},"d":"x,y"}`, 0, 2},
		{`x`, 0, -1},
		{`{"a":1`, 0, -1},
	}
	for _, tc := range objectCases {
		if got := countObjectFields([]byte(tc.data), tc.start); got != tc.want {
			t.Fatalf("countObjectFields(%q) = %d, want %d", tc.data, got, tc.want)
		}
	}

	arrayCases := []struct {
		data  string
		start int
		want  int
	}{
		{`[]`, 0, 0},
		{`[1]`, 0, 0},
		{`[1,[2,3],"a,b"]`, 0, 2},
		{`x`, 0, -1},
		{`[1,2`, 0, -1},
	}
	for _, tc := range arrayCases {
		if got := countArrayElements([]byte(tc.data), tc.start); got != tc.want {
			t.Fatalf("countArrayElements(%q) = %d, want %d", tc.data, got, tc.want)
		}
	}
}

func TestParserDoParseSpecificBranches(t *testing.T) {
	tests := []struct {
		data string
		kind core.NodeType
	}{
		{`{"a":1}`, core.Object},
		{`[1,2]`, core.Array},
		{`"x"`, core.String},
		{`true`, core.Bool},
		{`null`, core.Null},
		{`123`, core.Number},
	}
	for _, tc := range tests {
		p := newParser([]byte(tc.data), nil)
		node := p.doParse(nil)
		if node.Type() != tc.kind {
			t.Fatalf("doParse(%q) type=%v, want %v", tc.data, node.Type(), tc.kind)
		}
		p = newParser([]byte(tc.data), nil)
		node = p.doParseFull(nil)
		if node.Type() != tc.kind {
			t.Fatalf("doParseFull(%q) type=%v, want %v", tc.data, node.Type(), tc.kind)
		}
	}

	p := newParser([]byte(`?`), nil)
	if node := p.doParse(nil); node.Error() == nil {
		t.Fatal("expected doParse invalid char error")
	}
	p = newParser([]byte(`?`), nil)
	if node := p.doParseFull(nil); node.Error() == nil {
		t.Fatal("expected doParseFull invalid char error")
	}
}

func TestParserSkipWhitespace(t *testing.T) {
	p := newParser([]byte(" \n\r\tabc"), nil)
	p.skipWhitespace()
	if p.pos != 4 {
		t.Fatalf("expected pos 4, got %d", p.pos)
	}
}

func TestUnescapeWithBufferReuse(t *testing.T) {
	buf := make([]byte, 0, 32)
	first, err := unescapeWithBuffer([]byte(`hello`), &buf)
	if err != nil {
		t.Fatalf("first unescapeWithBuffer failed: %v", err)
	}
	second, err := unescapeWithBuffer([]byte(`x\ny`), &buf)
	if err != nil {
		t.Fatalf("second unescapeWithBuffer failed: %v", err)
	}
	_ = first
	if string(second) != "x\ny" {
		t.Fatalf("unexpected second output: %q", string(second))
	}
	if !bytes.Equal(second, []byte("x\ny")) {
		t.Fatalf("unexpected second bytes: %q", second)
	}
}