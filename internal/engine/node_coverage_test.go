package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Custom node that holds an unmarshalable type
type unmarshalableNode struct {
	baseNode
	value interface{}
}

// Implement the Node interface for unmarshalableNode
func (n *unmarshalableNode) Type() NodeType                            { return InvalidNode } // Mock type
func (n *unmarshalableNode) Get(key string) Node                       { return nil }
func (n *unmarshalableNode) Index(i int) Node                          { return nil }
func (n *unmarshalableNode) Query(path string) Node                    { return nil }
func (n *unmarshalableNode) ForEach(iterator func(interface{}, Node))  {}
func (n *unmarshalableNode) Len() int                                  { return 0 }
func (n *unmarshalableNode) String() string                            { return "" }
func (n *unmarshalableNode) MustString() string                        { return "" }
func (n *unmarshalableNode) Float() float64                            { return 0 }
func (n *unmarshalableNode) MustFloat() float64                        { return 0 }
func (n *unmarshalableNode) Int() int64                                { return 0 }
func (n *unmarshalableNode) MustInt() int64                            { return 0 }
func (n *unmarshalableNode) Bool() bool                                { return false }
func (n *unmarshalableNode) MustBool() bool                            { return false }
func (n *unmarshalableNode) Time() time.Time                           { return time.Time{} }
func (n *unmarshalableNode) MustTime() time.Time                       { return time.Time{} }
func (n *unmarshalableNode) Array() []Node                             { return nil }
func (n *unmarshalableNode) MustArray() []Node                         { return nil }
func (n *unmarshalableNode) AsMap() map[string]Node                    { return nil }
func (n *unmarshalableNode) MustAsMap() map[string]Node                { return nil }
func (n *unmarshalableNode) Interface() interface{}                    { return n.value }
func (n *unmarshalableNode) Func(name string, fn func(Node) Node) Node { return nil }
func (n *unmarshalableNode) CallFunc(name string) Node                 { return nil }
func (n *unmarshalableNode) RemoveFunc(name string) Node               { return nil }
func (n *unmarshalableNode) Filter(fn func(Node) bool) Node            { return nil }
func (n *unmarshalableNode) Map(fn func(Node) interface{}) Node        { return nil }
func (n *unmarshalableNode) Set(key string, value interface{}) Node    { return nil }
func (n *unmarshalableNode) Append(value interface{}) Node             { return nil }
func (n *unmarshalableNode) RawFloat() (float64, bool)                 { return 0, false }
func (n *unmarshalableNode) RawString() (string, bool)                 { return "", false }
func (n *unmarshalableNode) Contains(value string) bool                { return false }
func (n *unmarshalableNode) Strings() []string                         { return nil }

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

func TestNode_ForEachCoverage(t *testing.T) {
	t.Run("PrimitiveNodes", func(t *testing.T) {
		nodes := []Node{
			NewStringNode("test", "", nil),
			NewNumberNode(123, "", nil),
			NewBoolNode(true, "", nil),
			NewNullNode("", nil),
			NewInvalidNode("", errors.New("test error")),
		}

		for _, n := range nodes {
			t.Run(n.Type().String(), func(t *testing.T) {
				var called bool
				n.ForEach(func(i interface{}, n Node) {
					called = true
				})
				assert.False(t, called, "ForEach should be a no-op for %s", n.Type().String())
			})
		}
	})

	t.Run("ObjectNode", func(t *testing.T) {
		obj := NewObjectNode(map[string]Node{
			"a": NewStringNode("1", "", nil),
			"b": NewStringNode("2", "", nil),
		}, "", nil)

		count := 0
		obj.ForEach(func(key interface{}, value Node) {
			count++
			switch key.(string) {
			case "a":
				assert.Equal(t, "1", value.String())
			case "b":
				assert.Equal(t, "2", value.String())
			default:
				t.Errorf("unexpected key: %v", key)
			}
		})
		assert.Equal(t, 2, count)
	})

	t.Run("ArrayNode", func(t *testing.T) {
		arr := NewArrayNode([]Node{
			NewStringNode("a", "", nil),
			NewStringNode("b", "", nil),
		}, "", nil)

		count := 0
		arr.ForEach(func(index interface{}, value Node) {
			count++
			switch index.(int) {
			case 0:
				assert.Equal(t, "a", value.String())
			case 1:
				assert.Equal(t, "b", value.String())
			default:
				t.Errorf("unexpected index: %v", index)
			}
		})
		assert.Equal(t, 2, count)
	})
}

func TestNode_AppendCoverage(t *testing.T) {
	t.Run("Append to non-array nodes", func(t *testing.T) {
		nodes := []Node{
			NewObjectNode(map[string]Node{}, "", nil),
			NewStringNode("test", "root.string", nil),
			NewNumberNode(123, "root.number", nil),
			NewBoolNode(true, "root.bool", nil),
			NewNullNode("root.null", nil),
		}

		for _, n := range nodes {
			t.Run(n.Type().String(), func(t *testing.T) {
				res := n.Append("new value")
				assert.False(t, res.IsValid())
				assert.Equal(t, ErrTypeAssertion, res.Error())
			})
		}
	})

	t.Run("Append to manually constructed primitive root", func(t *testing.T) {
		nodes := []Node{
			NewStringNode("test", "", nil),
			NewNumberNode(123, "", nil),
			NewBoolNode(true, "", nil),
			NewNullNode("", nil),
		}

		for _, n := range nodes {
			t.Run(n.Type().String(), func(t *testing.T) {
				res := n.Append("new value")
				assert.False(t, res.IsValid())
				assert.Equal(t, ErrTypeAssertion, res.Error())
			})
		}
	})

	t.Run("Append with invalid value", func(t *testing.T) {
		arr := NewArrayNode([]Node{}, "", nil)
		res := arr.Append(make(chan int)) // Unsupported type
		assert.False(t, res.IsValid())
		assert.Error(t, res.Error())
	})
}

func TestNode_StringCoverage(t *testing.T) {
	t.Run("objectNode String with marshaling error", func(t *testing.T) {
		unmarshalable := &unmarshalableNode{value: make(chan int)}
		node := NewObjectNode(map[string]Node{"invalid": unmarshalable}, "", nil)

		str := node.String()
		assert.Equal(t, "", str)
		assert.Error(t, node.Error())
		assert.Contains(t, node.Error().Error(), "json: unsupported type")
	})

	t.Run("arrayNode String with marshaling error", func(t *testing.T) {
		unmarshalable := &unmarshalableNode{value: make(chan int)}
		node := NewArrayNode([]Node{unmarshalable}, "", nil)

		str := node.String()
		assert.Equal(t, "", str)
		assert.Error(t, node.Error())
		assert.Contains(t, node.Error().Error(), "json: unsupported type")
	})
}
