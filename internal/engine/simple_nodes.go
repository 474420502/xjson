package engine

import (
	"strconv"
	"time"

	"github.com/474420502/xjson/internal/core"
)

// stringNode represents a JSON string.
type stringNode struct {
	baseNode
	value string
}

func (n *stringNode) Type() core.NodeType {
	return core.String
}

// RawString returns the raw string value and true.
func (n *stringNode) RawString() (string, bool) {
	// TODO: Proper unescaping of JSON string
	return n.value, true
}

func (n *stringNode) String() string {
	return n.value
}

func (n *stringNode) MustString() string {
	return n.value
}

func (n *stringNode) Float() float64 {
	f, err := strconv.ParseFloat(n.value, 64)
	if err != nil {
		n.setError(err)
		return 0
	}
	return f
}

func (n *stringNode) MustFloat() float64 {
	f, err := strconv.ParseFloat(n.value, 64)
	if err != nil {
		panic(err)
	}
	return f
}

func (n *stringNode) Int() int64 {
	i, err := strconv.ParseInt(n.value, 10, 64)
	if err != nil {
		n.setError(err)
		return 0
	}
	return i
}

func (n *stringNode) MustInt() int64 {
	i, err := strconv.ParseInt(n.value, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func (n *stringNode) Bool() bool {
	b, err := strconv.ParseBool(n.value)
	if err != nil {
		n.setError(err)
		return false
	}
	return b
}

func (n *stringNode) MustBool() bool {
	b, err := strconv.ParseBool(n.value)
	if err != nil {
		panic(err)
	}
	return b
}

func (n *stringNode) Time() time.Time {
	// Assuming RFC3339 format, as it's common for JSON
	t, err := time.Parse(time.RFC3339, n.value)
	if err != nil {
		n.setError(err)
		return time.Time{}
	}
	return t
}

func (n *stringNode) MustTime() time.Time {
	t, err := time.Parse(time.RFC3339, n.value)
	if err != nil {
		panic(err)
	}
	return t
}

func (n *stringNode) Interface() interface{} {
	return n.value
}

func (n *stringNode) SetValue(value interface{}) core.Node {
	n.setError(core.ErrTypeAssertion) // Or a more specific error
	return n
}

// numberNode represents a JSON number.
type numberNode struct {
	baseNode
}

func (n *numberNode) Type() core.NodeType {
	return core.Number
}

func (n *numberNode) RawFloat() (float64, bool) {
	f, err := strconv.ParseFloat(n.Raw(), 64)
	return f, err == nil
}

func (n *numberNode) Float() float64 {
	f, _ := n.RawFloat()
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
	i, err := strconv.ParseInt(n.Raw(), 10, 64)
	if err != nil {
		// Try parsing as float first
		if f, ok := n.RawFloat(); ok {
			return int64(f)
		}
		n.setError(err)
		return 0
	}
	return i
}

func (n *numberNode) MustInt() int64 {
	i, err := strconv.ParseInt(n.Raw(), 10, 64)
	if err != nil {
		if f, ok := n.RawFloat(); ok {
			return int64(f)
		}
		panic(err)
	}
	return i
}

func (n *numberNode) Interface() interface{} {
	// Attempt to return the most precise type
	if i, err := strconv.ParseInt(n.Raw(), 10, 64); err == nil {
		return i
	}
	f, _ := n.RawFloat()
	return f
}

func (n *numberNode) SetValue(value interface{}) core.Node {
	n.setError(core.ErrTypeAssertion) // Or a more specific error
	return n
}

// boolNode represents a JSON boolean.
type boolNode struct {
	baseNode
	value bool
}

func (n *boolNode) Type() core.NodeType {
	return core.Bool
}

func (n *boolNode) Bool() bool {
	return n.value
}

func (n *boolNode) MustBool() bool {
	return n.value
}

func (n *boolNode) Interface() interface{} {
	return n.value
}

func (n *boolNode) SetValue(value interface{}) core.Node {
	n.setError(core.ErrTypeAssertion) // Or a more specific error
	return n
}

// nullNode represents a JSON null.
type nullNode struct {
	baseNode
}

func (n *nullNode) Type() core.NodeType {
	return core.Null
}

func (n *nullNode) Interface() interface{} {
	return nil
}

func (n *nullNode) SetValue(value interface{}) core.Node {
	n.setError(core.ErrTypeAssertion) // Or a more specific error
	return n
}
