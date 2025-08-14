package engine

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/474420502/xjson/internal/core"
)

type objectNode struct {
	baseNode
	value      map[string]core.Node
	sortedKeys []string
	isDirty    bool
}

func (n *objectNode) Type() core.NodeType {
	return core.Object
}

func (n *objectNode) Len() int {
	n.lazyParse()
	return len(n.value)
}

func (n *objectNode) Get(key string) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	
	if child, ok := n.value[key]; ok {
		return child
	}
	return newInvalidNode(fmt.Errorf("key not found: %s", key))
}

func (n *objectNode) ForEach(fn func(keyOrIndex interface{}, value core.Node)) {
	if n.err != nil {
		return
	}
	n.lazyParse()
	if n.sortedKeys == nil {
		keys := make([]string, 0, len(n.value))
		for k := range n.value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		n.sortedKeys = keys
	}

	for _, k := range n.sortedKeys {
		fn(k, n.value[k])
	}
}

func (n *objectNode) Set(key string, value interface{}) core.Node {
	if n.err != nil {
		return n
	}
	n.lazyParse()
	n.isDirty = true // Mark as dirty so String() will regenerate

	// Also mark all ancestors as dirty to ensure String() regeneration
	current := n.parent
	for current != nil {
		if obj, ok := current.(*objectNode); ok {
			obj.isDirty = true
			current = obj.parent
		} else if arr, ok := current.(*arrayNode); ok {
			arr.isDirty = true
			current = arr.parent
		} else {
			break
		}
	}
	
	// Clear query cache since we're modifying the node
	n.baseNode.clearQueryCache()

	// Update sorted keys
	found := false
	for _, k := range n.sortedKeys {
		if k == key {
			found = true
			break
		}
	}
	if !found {
		n.sortedKeys = append(n.sortedKeys, key)
		sort.Strings(n.sortedKeys)
	}

	child := NewNodeFromInterface(n, value, n.funcs)
	if !child.IsValid() {
		n.setError(child.Error())
		return n
	}
	n.value[key] = child

	return n
}

// SetByPath implements the SetByPath method for objectNode
func (n *objectNode) SetByPath(path string, value interface{}) core.Node {
	return n.baseNode.SetByPath(path, value)
}

// 新增辅助方法来避免重复代码
func (n *objectNode) containsKey(key string) bool {
	if n.sortedKeys == nil {
		return false
	}
	for _, k := range n.sortedKeys {
		if k == key {
			return true
		}
	}
	return false
}

func (n *objectNode) AsMap() map[string]core.Node {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	return n.value
}

func (n *objectNode) MustAsMap() map[string]core.Node {
	if n.err != nil {
		panic(n.err)
	}
	n.lazyParse()
	return n.value
}

func (n *objectNode) String() string {
	if n.err != nil {
		return ""
	}
	n.lazyParse()
	if !n.isDirty && n.Raw() != "" {
		// Check if any child node is dirty
		hasDirtyChild := false
		for _, child := range n.value {
			switch c := child.(type) {
			case *objectNode:
				if c.isDirty {
					hasDirtyChild = true
					break
				}
			case *arrayNode:
				if c.isDirty {
					hasDirtyChild = true
					break
				}
			}
		}
		
		if !hasDirtyChild {
			return n.Raw()
		}
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	if n.sortedKeys == nil {
		keys := make([]string, 0, len(n.value))
		for k := range n.value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		n.sortedKeys = keys
	}
	for i, k := range n.sortedKeys {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf("%q:%s", k, n.value[k].String()))
	}
	buf.WriteByte('}')
	return buf.String()
}

func (n *objectNode) Keys() []string {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	if n.sortedKeys == nil {
		keys := make([]string, 0, len(n.value))
		for k := range n.value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		n.sortedKeys = keys
	}
	return n.sortedKeys
}

func (n *objectNode) Interface() interface{} {
	if n.err != nil {
		return nil
	}
	n.lazyParse()
	m := make(map[string]interface{})
	for k, v := range n.value {
		m[k] = v.Interface()
	}
	return m
}

func (n *objectNode) lazyParse() {
	if n.parsed.Load() || n.isDirty {
		return
	}
	if len(n.raw) == 0 { // constructed node
		n.parsed.Store(true)
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.parsed.Load() || n.isDirty {
		return
	}
	defer n.parsed.Store(true)

	p := newParser(n.raw, n.funcs)
	// start from the beginning of raw to parse the object
	p.pos = 0
	// For root node, pass nil as parent to avoid setting root as its own parent
	var parent core.Node
	if n.parent != nil {
		parent = n
	}
	parsedNode := p.parseObject(parent)
	if err := parsedNode.Error(); err != nil {
		n.err = err
		return
	}

	// copy values
	if cast, ok := parsedNode.(*objectNode); ok {
		n.value = cast.value
		// Also copy sortedKeys to maintain order
		n.sortedKeys = cast.sortedKeys
	}
}

func (n *objectNode) addChild(key string, child core.Node) {
	if n.value == nil {
		n.value = make(map[string]core.Node)
	}
	
	// Instead of type assertion, we use the Parent() and SetParent() pattern
	// All node types embed baseNode which has parent field
	if bn, ok := child.(*baseNode); ok {
		bn.parent = n
	} else if inode, ok := child.(interface{ setParent(core.Node) }); ok {
		inode.setParent(n)
	}
	
	n.value[key] = child
}
