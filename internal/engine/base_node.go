package engine

import (
	"time"
)

// baseNode serves as the base for all concrete node types, implementing common
// functionalities like error handling and path management.
type baseNode struct {
	err   error
	path  string
	raw   *string
	funcs *map[string]func(Node) Node // Changed to pointer
}

func (n *baseNode) IsValid() bool {
	return n.err == nil
}

func (n *baseNode) Error() error {
	return n.err
}

func (n *baseNode) Path() string {
	return n.path
}

func (n *baseNode) Raw() string {
	if n.raw != nil {
		return *n.raw
	}
	return ""
}

func (n *baseNode) setError(err error) {
	if n.err == nil {
		n.err = err
	}
}

func (n *baseNode) GetFuncs() *map[string]func(Node) Node {
	return n.funcs
}

// Common methods that will be overridden by specific node types but need to be defined
// to satisfy the Node interface

func (n *baseNode) Type() NodeType                            { return InvalidNode }
func (n *baseNode) Get(key string) Node                       { return nil }
func (n *baseNode) Index(i int) Node                          { return nil }
func (n *baseNode) Query(path string) Node                    { return nil }
func (n *baseNode) ForEach(iterator func(interface{}, Node))  {}
func (n *baseNode) Len() int                                  { return 0 }
func (n *baseNode) String() string                            { return "" }
func (n *baseNode) MustString() string                        { panic("not implemented") }
func (n *baseNode) Float() float64                            { return 0 }
func (n *baseNode) MustFloat() float64                        { panic("not implemented") }
func (n *baseNode) Int() int64                                { return 0 }
func (n *baseNode) MustInt() int64                            { panic("not implemented") }
func (n *baseNode) Bool() bool                                { return false }
func (n *baseNode) MustBool() bool                            { panic("not implemented") }
func (n *baseNode) Time() time.Time                           { return time.Time{} }
func (n *baseNode) MustTime() time.Time                       { panic("not implemented") }
func (n *baseNode) Array() []Node                             { return nil }
func (n *baseNode) MustArray() []Node                         { panic("not implemented") }
func (n *baseNode) Interface() interface{}                    { return nil }
func (n *baseNode) Func(name string, fn func(Node) Node) Node { return nil }
func (n *baseNode) CallFunc(name string) Node                 { return nil }
func (n *baseNode) RemoveFunc(name string) Node               { return nil }
func (n *baseNode) Filter(fn func(Node) bool) Node            { return nil }
func (n *baseNode) Map(fn func(Node) interface{}) Node        { return nil }
func (n *baseNode) Set(key string, value interface{}) Node    { return nil }
func (n *baseNode) Append(value interface{}) Node             { return nil }
func (n *baseNode) RawFloat() (float64, bool)                 { return 0, false }
func (n *baseNode) RawString() (string, bool)                 { return "", false }
func (n *baseNode) Contains(value string) bool                { return false }
func (n *baseNode) Strings() []string                         { return nil }
func (n *baseNode) AsMap() map[string]Node                    { return nil }
func (n *baseNode) MustAsMap() map[string]Node                { panic("not implemented") }
