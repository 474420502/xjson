package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayNode_FuncCoverage(t *testing.T) {
	funcs := make(map[string]func(Node) Node)
	funcs["double"] = func(n Node) Node {
		return NewNumberNode(n.Float()*2, "", &funcs)
	}

	arr := NewArrayNode(
		[]Node{
			NewNumberNode(1, "", &funcs),
			NewNumberNode(2, "", &funcs),
		},
		"",
		&funcs,
	)

	t.Run("CallFunc", func(t *testing.T) {
		result := arr.CallFunc("double")
		assert.True(t, result.IsValid())
		assert.Equal(t, ArrayNode, result.Type())
		assert.Equal(t, 2, result.Len())
		assert.Equal(t, 2.0, result.Index(0).Float())
		assert.Equal(t, 4.0, result.Index(1).Float())
	})

	t.Run("RemoveFunc", func(t *testing.T) {
		arr.RemoveFunc("double")
		result := arr.CallFunc("double")
		assert.False(t, result.IsValid())
	})
}
