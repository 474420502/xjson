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
	assert.True(t, root.IsValid())

	// Initially no funcs (implementation may return empty map or nil)
	initialFuncs := root.GetFuncs()
	if initialFuncs != nil {
		assert.Equal(t, 0, len(*initialFuncs))
	}

	// Register a function
	root = root.Func("double", func(n Node) Node {
		v := n.Get("value")
		if v.IsValid() && v.Type() == NumberNode {
			return n.Set("value", v.Float()*2)
		}
		return n
	})

	// Ensure CallFunc works
	res := root.CallFunc("double")
	assert.True(t, res.IsValid())
	assert.Equal(t, 2.0, res.Get("value").Float())

	// GetFuncs should list our function
	funcs := root.GetFuncs()
	assert.NotNil(t, funcs)
	assert.Contains(t, *funcs, "double")

	// Invoke function from returned map directly (tests wrapper translation)
	f := (*funcs)["double"]
	res2 := f(root)
	assert.True(t, res2.IsValid())
	assert.Equal(t, 4.0, res2.Get("value").Float())

	// Mutate returned map (should not affect internal funcs)
	delete(*funcs, "double")
	// Original CallFunc should still work (internal map intact)
	res3 := root.CallFunc("double")
	assert.True(t, res3.IsValid())
	assert.Equal(t, 8.0, res3.Get("value").Float())
}
