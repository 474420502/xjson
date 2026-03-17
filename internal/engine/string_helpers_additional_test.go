package engine

import "testing"

func TestStringNodeRawStringBranches(t *testing.T) {
	decoded := NewDecodedStringNode(nil, []byte("ready"), nil).(*stringNode)
	if got, ok := decoded.RawString(); !ok || got != "ready" {
		t.Fatalf("unexpected decoded RawString: %q %v", got, ok)
	}

	zeroCopy := NewRawStringNode(nil, []byte(`"hello"`), 1, 6, false, nil).(*stringNode)
	if got, ok := zeroCopy.RawString(); !ok || got != "hello" {
		t.Fatalf("unexpected zero-copy RawString: %q %v", got, ok)
	}

	escaped := NewRawStringNode(nil, []byte(`"a\nb"`), 1, 5, true, nil).(*stringNode)
	if got, ok := escaped.RawString(); !ok || got != "a\nb" {
		t.Fatalf("unexpected escaped RawString: %q %v", got, ok)
	}

	invalidBounds := NewRawStringNode(nil, []byte(`"x"`), 3, 1, false, nil).(*stringNode)
	if got, ok := invalidBounds.RawString(); ok || got != "" {
		t.Fatalf("expected invalid bounds RawString failure, got %q %v", got, ok)
	}

	badEscape := NewRawStringNode(nil, []byte{'"', '\\'}, 1, 2, true, nil).(*stringNode)
	if got, ok := badEscape.RawString(); ok || got != "" || badEscape.Error() == nil {
		t.Fatalf("expected bad escape RawString failure, got %q %v err=%v", got, ok, badEscape.Error())
	}
	if !badEscape.Contains("anything") && badEscape.Error() == nil {
		t.Fatal("expected Contains to preserve error state usage")
	}
}