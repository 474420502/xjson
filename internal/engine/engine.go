package engine

import (
	"errors"
)

// Note: The local Node interface has been removed.
// All node implementations should now directly use the `core.Node` interface
// from `github.com/474420502/xjson/internal/core`.
//
// The NodeType enum has also been centralized in the core package.

var (
	ErrInvalidNode      = errors.New("invalid node")
	ErrTypeAssertion    = errors.New("type assertion failed")
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrNotFound         = errors.New("not found")
)
