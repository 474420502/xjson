package engine

import (
	"testing"
)

func TestSimpleMustBool(t *testing.T) {
	strNode := NewStringNode("hello", "", nil)
	defer func() {
		if r := recover(); r != nil {
			t.Log("MustBool panicked as expected")
		} else {
			t.Error("MustBool did not panic")
		}
	}()

	strNode.MustBool()
	t.Error("Should have panicked")
}
