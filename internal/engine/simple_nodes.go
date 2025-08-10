package engine

import (
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

func (n *numberNode) RawFloat() (float64, bool) {
	f, err := strconv.ParseFloat(n.Raw(), 64)
	if err != nil {
		return 0, false
	}
	return f, true
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

// nullNode implementation
type nullNode struct {
	baseNode
}

func (n *nullNode) Type() core.NodeType    { return core.Null }
func (n *nullNode) String() string         { return "null" }
func (n *nullNode) Interface() interface{} { return nil }
