package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayAppendAndString(t *testing.T) {
	root, err := Parse([]byte(`{"a": []}`))
	assert.NoError(t, err)
	arr := root.Query("/a")
	assert.True(t, arr.IsValid())
	arr.Append(1).Append(2)
	assert.Equal(t, 2, arr.Len())
	// ensure printable
	s := root.String()
	assert.Contains(t, s, "[1,2]")
}

func TestKeysSorted(t *testing.T) {
	root, err := Parse([]byte(`{"b":1,"a":2}`))
	assert.NoError(t, err)
	keys := root.(*objectNode).Keys()
	assert.Equal(t, []string{"a", "b"}, keys)
}

func TestNumberAndStringRawAccess(t *testing.T) {
	root, err := Parse([]byte(`{"n":12.34,"i":5,"s":"hello"}`))
	assert.NoError(t, err)
	n := root.Query("/n")
	f, ok := n.RawFloat()
	assert.True(t, ok)
	assert.InDelta(t, 12.34, f, 1e-9)

	i := root.Query("/i")
	fi, ok := i.RawFloat()
	assert.True(t, ok)
	assert.InDelta(t, 5.0, fi, 1e-9)

	s := root.Query("/s")
	str, ok := s.RawString()
	assert.True(t, ok)
	assert.Equal(t, "hello", str)
}

func TestStringsHelper(t *testing.T) {
	root, err := Parse([]byte(`{"arr":["a","b","c"]}`))
	assert.NoError(t, err)
	ss := root.Query("/arr").Strings()
	assert.Equal(t, []string{"a", "b", "c"}, ss)
}
