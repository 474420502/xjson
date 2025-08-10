package engine

import (
	"fmt"

	"github.com/474420502/xjson/internal/core"
)

// slice represents a slice operation, e.g., [start:end].
type slice struct {
	Start, End int
}

// ParseQuery tokenizes a query path into a sequence of operations.
func ParseQuery(path string) ([]core.QueryToken, error) {
	// TODO: This is the core of the query parser. It needs to handle all the
	// syntax described in the README (keys, indexes, funcs, wildcards, etc.).
	// For now, it will return an empty list.
	return nil, fmt.Errorf("ParseQuery not implemented")
}

func (n *objectNode) Query(path string) core.Node {
	// TODO: Implement query logic for object nodes
	return newInvalidNodeWithMsg("Query on object not yet implemented: %s", path)
}

func (n *arrayNode) Query(path string) core.Node {
	// TODO: Implement query logic for array nodes
	return newInvalidNodeWithMsg("Query on array not yet implemented: %s", path)
}
