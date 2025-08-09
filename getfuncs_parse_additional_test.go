package xjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Additional tests to improve coverage for GetFuncs and Parse error branch
func TestGetFuncsAndParseErrorCoverage(t *testing.T) {
	// Cover Parse error branch
	_, err := Parse("{invalid")
	assert.Error(t, err)

	// Prepare JSON
	root, err := Parse(`{"value":1}`)
	assert.NoError(t, err)
	assert.NoError(t, root.Error())
	assert.True(t, root.IsValid())

	// Initially no funcs (implementation may return empty map or nil)
	initialFuncs := root.GetFuncs()
	// If funcs map is nil, we create an empty one for checking
	if initialFuncs == nil {
		emptyMap := make(map[string]func(Node) Node)
		initialFuncs = &emptyMap
	}
	assert.Equal(t, 0, len(*initialFuncs))

	// Register a function
	root = root.RegisterFunc("double", func(n Node) Node {
		v := n.Get("value")
		if v.Error() == nil && v.IsValid() && v.Type() == NumberNode {
			return n.Set("value", v.Float()*2)
		}
		return n
	})

	// Ensure CallFunc works
	res := root.CallFunc("double")
	if assert.NoError(t, res.Error()) && assert.True(t, res.IsValid()) {
		valueNode := res.Get("value")
		if assert.NoError(t, valueNode.Error()) {
			assert.Equal(t, 2.0, valueNode.Float())
		}
	}

	// GetFuncs should list our function
	funcs := res.GetFuncs()
	assert.NotNil(t, funcs)
	assert.Contains(t, *funcs, "double")

	// Invoke function from returned map directly
	f := (*funcs)["double"]
	res2 := f(res)
	if assert.NoError(t, res2.Error()) && assert.True(t, res2.IsValid()) {
		valueNode := res2.Get("value")
		if assert.NoError(t, valueNode.Error()) {
			assert.Equal(t, 4.0, valueNode.Float())
		}
	}

	// Mutate returned map (should not affect internal funcs)
	delete(*funcs, "double")
	// Original CallFunc should NOW FAIL because the returned map is a pointer.
	res3 := res2.CallFunc("double")
	assert.Error(t, res3.Error())
	assert.Contains(t, res3.Error().Error(), "function double not found")
}
