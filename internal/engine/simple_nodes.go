package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/474420502/xjson/internal/core"
)

// stringNode implementation
type stringNode struct {
	baseNode
	value string
}

func (n *stringNode) Type() core.NodeType { return core.String }
func (n *stringNode) String() string      { return n.value }
func (n *stringNode) MustString() string  { return n.value }
func (n *stringNode) RawString() (string, bool) {
	return n.value, true
}
func (n *stringNode) Contains(v string) bool { return n.value == v }
func (n *stringNode) Interface() interface{} {
	return n.value
}

func (n *stringNode) Set(key string, value interface{}) core.Node {
	return newInvalidNode(fmt.Errorf("set not supported on type %s", n.Type()))
}

// SetByPath implements the SetByPath method for stringNode
func (n *stringNode) SetByPath(path string, value interface{}) core.Node {
	return n.baseNode.SetByPath(path, value)
}

func (n *stringNode) Time() time.Time {
	t, err := time.Parse(time.RFC3339Nano, n.value)
	if err != nil {
		n.setError(err)
		return time.Time{}
	}
	return t
}

func (n *stringNode) MustTime() time.Time {
	t, err := time.Parse(time.RFC3339Nano, n.value)
	if err != nil {
		panic(err)
	}
	return t
}

// numberNode implementation
type numberNode struct {
	baseNode
}

func (n *numberNode) Type() core.NodeType { return core.Number }

func (n *numberNode) Float() float64 {
	f, _ := strconv.ParseFloat(n.Raw(), 64)
	return f
}

func (n *numberNode) MustFloat() float64 {
	f, err := strconv.ParseFloat(n.Raw(), 64)
	if err != nil {
		panic(err)
	}
	return f
}

func (n *numberNode) Int() int64 {
	i, _ := strconv.ParseInt(n.Raw(), 10, 64)
	return i
}

func (n *numberNode) MustInt() int64 {
	i, err := strconv.ParseInt(n.Raw(), 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func (n *numberNode) Interface() interface{} {
	raw := n.Raw()
	if !strings.Contains(raw, ".") {
		if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return i
		}
	}
	f, _ := strconv.ParseFloat(raw, 64)
	return f
}

func (n *numberNode) Set(key string, value interface{}) core.Node {
	return newInvalidNode(fmt.Errorf("set not supported on type %s", n.Type()))
}

// SetByPath implements the SetByPath method for numberNode
func (n *numberNode) SetByPath(path string, value interface{}) core.Node {
	return n.baseNode.SetByPath(path, value)
}

func (n *numberNode) RawFloat() (float64, bool) {
	f, err := strconv.ParseFloat(n.Raw(), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func (n *numberNode) RawString() (string, bool) {
	return n.Raw(), true
}

// boolNode implementation
type boolNode struct {
	baseNode
	value bool
}

func (n *boolNode) Type() core.NodeType { return core.Bool }
func (n *boolNode) Bool() bool          { return n.value }
func (n *boolNode) MustBool() bool      { return n.value }
func (n *boolNode) String() string {
	if n.value {
		return "true"
	}
	return "false"
}
func (n *boolNode) Interface() interface{} { return n.value }

func (n *boolNode) RawString() (string, bool) {
	if n.value {
		return "true", true
	}
	return "false", true
}

func (n *boolNode) RawFloat() (float64, bool) {
	return 0, false
}

func (n *boolNode) Set(key string, value interface{}) core.Node {
	return newInvalidNode(fmt.Errorf("set not supported on type %s", n.Type()))
}

// SetByPath implements the SetByPath method for boolNode
func (n *boolNode) SetByPath(path string, value interface{}) core.Node {
	return n.baseNode.SetByPath(path, value)
}

// nullNode implementation
type nullNode struct {
	baseNode
}

func (n *nullNode) Type() core.NodeType    { return core.Null }
func (n *nullNode) String() string         { return "null" }
func (n *nullNode) Interface() interface{} { return nil }

func (n *nullNode) RawString() (string, bool) {
	return "null", true
}

func (n *nullNode) RawFloat() (float64, bool) {
	return 0, false
}

func (n *nullNode) Set(key string, value interface{}) core.Node {
	return newInvalidNode(fmt.Errorf("set not supported on type %s", n.Type()))
}

// SetByPath implements the SetByPath method for nullNode
func (n *nullNode) SetByPath(path string, value interface{}) core.Node {
	return n.baseNode.SetByPath(path, value)
}
