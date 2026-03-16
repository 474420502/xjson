package query

// Op represents a query operation type.
type Op int

const (
	OpKey Op = iota
	OpIndex
	OpSlice
	OpFunc
	OpWildcard
	OpRecursiveKey
	OpParent
	OpAll
)

// QueryToken represents a single token in a parsed query.
type QueryToken struct {
	Type  Op
	Value interface{}
}
