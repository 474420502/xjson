package query

// Query represents a parsed path
type Query struct {
	Segments []Segment
}

// Segment represents a single part of a query path
type Segment struct {
	Type        SegmentType
	Value       string
	Slice       *Slice
	Filter      *Filter
	IsRecursive bool
}

// SegmentType is the type of a path segment
type SegmentType int

const (
	FieldSegment SegmentType = iota
	IndexSegment
	SliceSegment
	WildcardSegment
	FilterSegment
)

// Slice represents an array slice operation
type Slice struct {
	Start *int
	End   *int
}

// Filter represents a filter operation
type Filter struct {
	Expression string
}
