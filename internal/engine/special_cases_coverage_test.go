package engine

import (
	"errors"
	"strings"
	"testing"

	"github.com/474420502/xjson/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestBoolNodeApply_SpecialCases(t *testing.T) {
	funcs := &map[string]func(core.Node) core.Node{}
	bn := NewBoolNode(true, "", funcs)

	// nil fn must panic
	assert.Panics(t, func() { bn.Apply(nil) })

	// predicate true -> returns same node
	res := bn.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
	assert.Same(t, bn, res)
	assert.Equal(t, core.BoolNode, res.Type())

	// predicate false -> invalid node with error
	res = bn.Apply(core.PredicateFunc(func(n core.Node) bool { return false }))
	assert.Equal(t, core.InvalidNode, res.Type())
	if err := res.Error(); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "predicate returned false")
	}

	// transform -> new node (object)
	res = bn.Apply(core.TransformFunc(func(n core.Node) interface{} {
		return map[string]interface{}{"x": 1}
	}))
	assert.Equal(t, core.ObjectNode, res.Type())

	// transform -> unsupported type -> invalid
	res = bn.Apply(core.TransformFunc(func(n core.Node) interface{} {
		return make(chan int)
	}))
	assert.Equal(t, core.InvalidNode, res.Type())
	assert.Error(t, res.Error())

	// unsupported signature -> invalid
	res = bn.Apply(core.UnaryPathFunc(func(n core.Node) core.Node { return n }))
	assert.Equal(t, core.InvalidNode, res.Type())
	if err := res.Error(); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unsupported function signature")
	}

	// node with internal error returns itself
	errNode := &boolNode{baseNode: baseNode{err: errors.New("boom")}}
	res = errNode.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
	assert.Same(t, core.Node(errNode), res)
}

func TestFactory_NewNodeFromInterface_ArrayBranch(t *testing.T) {
	funcs := &map[string]func(core.Node) core.Node{}
	input := []interface{}{1, "two", true, nil, map[string]interface{}{"a": 1}, []interface{}{2, 3}}

	n, err := NewNodeFromInterface(input, "", funcs)
	assert.NoError(t, err)
	assert.Equal(t, core.ArrayNode, n.Type())
	assert.Equal(t, 6, n.Len())

	// check element types
	assert.Equal(t, core.NumberNode, n.Index(0).Type())
	assert.Equal(t, core.StringNode, n.Index(1).Type())
	assert.Equal(t, core.BoolNode, n.Index(2).Type())
	assert.Equal(t, core.NullNode, n.Index(3).Type())

	// object element and nested path correctness
	objEl := n.Index(4)
	assert.Equal(t, core.ObjectNode, objEl.Type())
	av := objEl.Get("a")
	assert.Equal(t, core.NumberNode, av.Type())
	assert.Equal(t, "[4].a", av.Path())

	// nested array element and child path correctness
	arrEl := n.Index(5)
	assert.Equal(t, core.ArrayNode, arrEl.Type())
	assert.Equal(t, 2, arrEl.Len())
	firstInner := arrEl.Index(0)
	assert.Equal(t, core.NumberNode, firstInner.Type())
	assert.Equal(t, "[5][0]", firstInner.Path())
}

func TestEvaluateQuery_WildcardOnObject(t *testing.T) {
	funcs := &map[string]func(core.Node) core.Node{}
	m := map[string]core.Node{
		"a": NewNumberNode(1, "", funcs),
		"b": NewStringNode("x", "", funcs),
		"c": NewBoolNode(true, "", funcs),
	}
	obj := NewObjectNode(m, "", funcs)

	res := obj.Query("*")
	assert.Equal(t, core.ArrayNode, res.Type())
	assert.Equal(t, 3, res.Len())

	// all elements should be valid
	var validCount int
	res.ForEach(func(_ interface{}, v core.Node) {
		if v.IsValid() {
			validCount++
		}
	})
	assert.Equal(t, 3, validCount)

	// wildcard not applicable after selecting a scalar: go through EvaluateQuery via object.Query("a/*")
	invalidRes := obj.Query("a/*")
	assert.Equal(t, core.InvalidNode, invalidRes.Type())
	if err := invalidRes.Error(); assert.Error(t, err) {
		assert.True(t, strings.Contains(err.Error(), "wildcard"))
	}
}
