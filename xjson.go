// Package xjson provides a high-performance JSON library with XPath-like query support,
// lazy parsing, and seamless read-write operations.
//
// XJSON solves the performance vs flexibility gap between gjson (fast but read-only)
// and the standard library (flexible but slow) by combining:
//   - XPath-like query syntax for powerful JSON traversal
//   - Lazy parsing for zero-allocation read operations
//   - Materialize-on-write for seamless mutation support
//   - Interface-driven design for cross-language portability
package xjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/474420502/xjson/internal/engine"
	"github.com/474420502/xjson/internal/modifier"
	"github.com/474420502/xjson/internal/parser"
)

// Common errors
var (
	ErrInvalidJSON    = errors.New("invalid JSON")
	ErrInvalidPath    = errors.New("invalid path")
	ErrNotFound       = errors.New("path not found")
	ErrTypeMismatch   = errors.New("type mismatch")
	ErrReadOnlyResult = errors.New("result is read-only")
)

// IDocument defines the interface for JSON document operations
type IDocument interface {
	// Query executes an XPath-like query and returns matching results
	Query(xpath string) IResult

	// Set modifies a value at the specified path (triggers materialization)
	Set(path string, value interface{}) error

	// Delete removes a value at the specified path (triggers materialization)
	Delete(path string) error

	// Bytes returns the JSON representation as bytes
	Bytes() ([]byte, error)

	// String returns the JSON representation as string
	String() (string, error)

	// IsValid checks if the document contains valid JSON
	IsValid() bool

	// IsMaterialized returns true if the document has been materialized
	IsMaterialized() bool
}

// IResult defines the interface for query results
type IResult interface {
	// Type conversion methods
	String() (string, error)
	MustString() string
	Int() (int, error)
	MustInt() int
	Int64() (int64, error)
	MustInt64() int64
	Float() (float64, error)
	MustFloat() float64
	Bool() (bool, error)
	MustBool() bool

	// Array/Object methods
	Get(path string) IResult
	Index(i int) IResult
	Count() int
	Keys() []string

	// Iteration methods
	ForEach(func(index int, value IResult) bool)
	Map(func(index int, value IResult) interface{}) []interface{}
	Filter(func(index int, value IResult) bool) IResult

	// Utility methods
	Exists() bool
	IsNull() bool
	IsArray() bool
	IsObject() bool
	First() IResult
	Last() IResult

	// Raw access
	Raw() interface{}
	Bytes() ([]byte, error)
}

// Document represents a JSON document with lazy parsing and materialize-on-write capabilities
type Document struct {
	// Raw JSON bytes (nil after materialization)
	raw []byte

	// Materialized Go structure (nil before first write operation)
	materialized interface{}

	// State flags
	isMaterialized bool
	isValid        bool

	// Error state
	err error

	// Modifier for write operations
	mod *modifier.Modifier
}

// Parse creates a new Document from JSON bytes using lazy parsing
func Parse(data []byte) (*Document, error) {
	doc := &Document{
		raw:     make([]byte, len(data)),
		isValid: json.Valid(data),
	}

	copy(doc.raw, data)

	if !doc.isValid {
		doc.err = ErrInvalidJSON
		return doc, doc.err
	}

	return doc, nil
}

// ParseString creates a new Document from JSON string
func ParseString(data string) (*Document, error) {
	return Parse([]byte(data))
}

// Query executes an XPath-like query on the document
func (d *Document) Query(xpath string) IResult {
	if d.err != nil {
		return &Result{err: d.err}
	}

	if !d.isValid {
		return &Result{err: ErrInvalidJSON}
	}

	// For now, support simple dot-notation paths
	// Full XPath parsing will be implemented later
	if isSimplePath(xpath) {
		return d.querySimplePath(xpath)
	}

	// Try to parse as XPath query
	p := parser.NewParser(xpath)
	query, err := p.Parse()
	if err != nil {
		return &Result{err: fmt.Errorf("invalid query syntax: %w", err)}
	}

	var matches []engine.Match
	var queryErr error

	if d.isMaterialized {
		// Query on materialized data
		eng := engine.NewEngine()
		matches, queryErr = eng.ExecuteOnMaterialized(d.materialized, query)
	} else {
		// Query on raw bytes
		eng := engine.NewEngine()
		matches, queryErr = eng.ExecuteOnRaw(d.raw, query)
	}

	if queryErr != nil {
		return &Result{err: queryErr}
	}

	return &Result{
		doc:     d,
		matches: convertMatches(matches),
	}
}

// querySimplePath handles simple dot-notation paths like "store.book[0].title"
func (d *Document) querySimplePath(path string) IResult {
	if d.isMaterialized {
		// Query on materialized data
		if value, exists := engine.GetValueBySimplePath(d.materialized, path); exists {
			return &Result{
				doc:     d,
				matches: []interface{}{value},
			}
		}
		return &Result{doc: d, matches: []interface{}{}}
	}

	// Query on raw bytes using lazy parsing
	// This should NOT trigger materialization
	if value, exists := engine.GetValueBySimplePathFromRaw(d.raw, path); exists {
		return &Result{
			doc:     d,
			matches: []interface{}{value},
		}
	}

	return &Result{doc: d, matches: []interface{}{}}
}

// isSimplePath checks if the path is a simple dot notation
func isSimplePath(path string) bool {
	// Simple heuristic: if it doesn't contain XPath special characters, treat as simple path
	return !strings.Contains(path, "//") &&
		!strings.Contains(path, "[?") &&
		!strings.Contains(path, "@") &&
		!strings.Contains(path, "*")
}

// convertMatches converts engine matches to interface{} slice
func convertMatches(matches []engine.Match) []interface{} {
	result := make([]interface{}, len(matches))
	for i, match := range matches {
		result[i] = match.Value
	}
	return result
}

// Set modifies a value at the specified path, triggering materialization if needed
func (d *Document) Set(path string, value interface{}) error {
	if d.err != nil {
		return d.err
	}

	// Trigger materialization if not already done
	if !d.isMaterialized {
		if err := d.materialize(); err != nil {
			return err
		}
	}

	// Initialize modifier if not yet created
	if d.mod == nil {
		d.mod = modifier.NewModifier()
	}

	// Use modifier to set the value
	return d.mod.Set(&d.materialized, path, value)
}

// Delete removes a value at the specified path, triggering materialization if needed
func (d *Document) Delete(path string) error {
	if d.err != nil {
		return d.err
	}

	// Trigger materialization if not already done
	if !d.isMaterialized {
		if err := d.materialize(); err != nil {
			return err
		}
	}

	// Initialize modifier if not yet created
	if d.mod == nil {
		d.mod = modifier.NewModifier()
	}

	// Use modifier to delete the value
	return d.mod.Delete(&d.materialized, path)
}

// Bytes returns the JSON representation as bytes
func (d *Document) Bytes() ([]byte, error) {
	if d.err != nil {
		return nil, d.err
	}

	if d.isMaterialized {
		return json.Marshal(d.materialized)
	}

	return d.raw, nil
}

// String returns the JSON representation as string
func (d *Document) String() (string, error) {
	bytes, err := d.Bytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// IsValid checks if the document contains valid JSON
func (d *Document) IsValid() bool {
	return d.isValid && d.err == nil
}

// IsMaterialized returns true if the document has been materialized
func (d *Document) IsMaterialized() bool {
	return d.isMaterialized
}

// materialize converts raw JSON bytes to Go structures for write operations
func (d *Document) materialize() error {
	if d.isMaterialized {
		return nil
	}

	if !d.isValid {
		return ErrInvalidJSON
	}

	err := json.Unmarshal(d.raw, &d.materialized)
	if err != nil {
		d.err = err
		return err
	}

	d.isMaterialized = true
	d.raw = nil // Release raw bytes to save memory

	return nil
}

// Result represents the result of a query operation
type Result struct {
	doc     *Document
	matches []interface{} // Internal representation of matches
	err     error
}

// Implement IResult interface methods
// Note: These are placeholder implementations

func (r *Result) String() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if len(r.matches) == 0 {
		return "", ErrNotFound
	}

	value := r.matches[0]
	if value == nil {
		return "", nil
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		// For complex types, return JSON representation
		if bytes, err := json.Marshal(value); err == nil {
			return string(bytes), nil
		}
		return fmt.Sprintf("%v", value), nil
	}
}

func (r *Result) MustString() string {
	s, err := r.String()
	if err != nil {
		panic(err)
	}
	return s
}

func (r *Result) Int() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(r.matches) == 0 {
		return 0, ErrNotFound
	}

	value := r.matches[0]
	if value == nil {
		return 0, ErrTypeMismatch
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
		return 0, ErrTypeMismatch
	default:
		return 0, ErrTypeMismatch
	}
}

func (r *Result) MustInt() int {
	i, err := r.Int()
	if err != nil {
		panic(err)
	}
	return i
}

func (r *Result) Int64() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(r.matches) == 0 {
		return 0, ErrNotFound
	}

	value := r.matches[0]
	if value == nil {
		return 0, ErrTypeMismatch
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, nil
		}
		return 0, ErrTypeMismatch
	default:
		return 0, ErrTypeMismatch
	}
}

func (r *Result) MustInt64() int64 {
	i, err := r.Int64()
	if err != nil {
		panic(err)
	}
	return i
}

func (r *Result) Float() (float64, error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(r.matches) == 0 {
		return 0, ErrNotFound
	}

	value := r.matches[0]
	if value == nil {
		return 0, ErrTypeMismatch
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
		return 0, ErrTypeMismatch
	default:
		return 0, ErrTypeMismatch
	}
}

func (r *Result) MustFloat() float64 {
	f, err := r.Float()
	if err != nil {
		panic(err)
	}
	return f
}

func (r *Result) Bool() (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	if len(r.matches) == 0 {
		return false, ErrNotFound
	}

	value := r.matches[0]
	if value == nil {
		return false, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, nil
		}
		// Non-empty strings are truthy
		return v != "", nil
	case float64:
		return v != 0, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	default:
		// Non-nil values are truthy
		return true, nil
	}
}

func (r *Result) MustBool() bool {
	b, err := r.Bool()
	if err != nil {
		panic(err)
	}
	return b
}

func (r *Result) Get(path string) IResult {
	if r.err != nil {
		return &Result{err: r.err}
	}
	if len(r.matches) == 0 {
		return &Result{err: ErrNotFound}
	}

	// Get sub-path from the first match
	firstMatch := r.matches[0]
	if value, exists := engine.GetValueBySimplePath(firstMatch, path); exists {
		return &Result{
			doc:     r.doc,
			matches: []interface{}{value},
		}
	}

	return &Result{
		doc:     r.doc,
		matches: []interface{}{},
	}
}

func (r *Result) Index(i int) IResult {
	if r.err != nil {
		return &Result{err: r.err}
	}
	if len(r.matches) == 0 {
		return &Result{err: ErrNotFound}
	}

	// If we have multiple matches, return the i-th match directly
	if len(r.matches) > 1 {
		if i < 0 {
			i = len(r.matches) + i // Handle negative indices
		}
		if i >= 0 && i < len(r.matches) {
			return &Result{
				doc:     r.doc,
				matches: []interface{}{r.matches[i]},
			}
		}
		return &Result{err: ErrNotFound}
	}

	// Single match - try to access array index from first match
	firstMatch := r.matches[0]
	switch v := firstMatch.(type) {
	case []interface{}:
		if i < 0 {
			i = len(v) + i // Handle negative indices
		}
		if i >= 0 && i < len(v) {
			return &Result{
				doc:     r.doc,
				matches: []interface{}{v[i]},
			}
		}
		return &Result{
			doc:     r.doc,
			matches: []interface{}{},
		}
	default:
		return &Result{err: ErrTypeMismatch}
	}
}

func (r *Result) Count() int {
	if r.err != nil {
		return 0
	}

	if len(r.matches) == 0 {
		return 0
	}

	// If we have exactly one match and it's an array, return the array length
	if len(r.matches) == 1 {
		if arr, ok := r.matches[0].([]interface{}); ok {
			return len(arr)
		}
	}

	// Otherwise return the number of matches
	return len(r.matches)
}

func (r *Result) Keys() []string {
	if r.err != nil || len(r.matches) == 0 {
		return nil
	}

	firstMatch := r.matches[0]
	switch v := firstMatch.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		return keys
	default:
		return nil
	}
}

func (r *Result) ForEach(fn func(index int, value IResult) bool) {
	if r.err != nil {
		return
	}

	// If we have exactly one match and it's an array, iterate over the array elements
	if len(r.matches) == 1 {
		if arr, ok := r.matches[0].([]interface{}); ok {
			for i, item := range arr {
				result := &Result{
					doc:     r.doc,
					matches: []interface{}{item},
				}
				if !fn(i, result) {
					break
				}
			}
			return
		}
	}

	// Otherwise iterate over matches
	for i, match := range r.matches {
		result := &Result{
			doc:     r.doc,
			matches: []interface{}{match},
		}
		if !fn(i, result) {
			break
		}
	}
}

func (r *Result) Map(fn func(index int, value IResult) interface{}) []interface{} {
	if r.err != nil {
		return nil
	}

	// If we have exactly one match and it's an array, map over the array elements
	if len(r.matches) == 1 {
		if arr, ok := r.matches[0].([]interface{}); ok {
			results := make([]interface{}, len(arr))
			for i, item := range arr {
				result := &Result{
					doc:     r.doc,
					matches: []interface{}{item},
				}
				results[i] = fn(i, result)
			}
			return results
		}
	}

	// Otherwise map over matches
	results := make([]interface{}, len(r.matches))
	for i, match := range r.matches {
		result := &Result{
			doc:     r.doc,
			matches: []interface{}{match},
		}
		results[i] = fn(i, result)
	}
	return results
}

func (r *Result) Filter(fn func(index int, value IResult) bool) IResult {
	if r.err != nil {
		return &Result{err: r.err}
	}

	var filtered []interface{}
	for i, match := range r.matches {
		result := &Result{
			doc:     r.doc,
			matches: []interface{}{match},
		}
		if fn(i, result) {
			filtered = append(filtered, match)
		}
	}

	return &Result{
		doc:     r.doc,
		matches: filtered,
	}
}

func (r *Result) Exists() bool {
	return r.err == nil && len(r.matches) > 0
}

func (r *Result) IsNull() bool {
	if r.err != nil || len(r.matches) == 0 {
		return false
	}
	return r.matches[0] == nil
}

func (r *Result) IsArray() bool {
	if r.err != nil || len(r.matches) == 0 {
		return false
	}
	_, ok := r.matches[0].([]interface{})
	return ok
}

func (r *Result) IsObject() bool {
	if r.err != nil || len(r.matches) == 0 {
		return false
	}
	_, ok := r.matches[0].(map[string]interface{})
	return ok
}

func (r *Result) First() IResult {
	if r.err != nil || len(r.matches) == 0 {
		return &Result{err: ErrNotFound}
	}
	return &Result{
		doc:     r.doc,
		matches: []interface{}{r.matches[0]},
	}
}

func (r *Result) Last() IResult {
	if r.err != nil || len(r.matches) == 0 {
		return &Result{err: ErrNotFound}
	}
	return &Result{
		doc:     r.doc,
		matches: []interface{}{r.matches[len(r.matches)-1]},
	}
}

func (r *Result) Raw() interface{} {
	if r.err != nil || len(r.matches) == 0 {
		return nil
	}
	return r.matches[0]
}

func (r *Result) Bytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.matches) == 0 {
		return nil, ErrNotFound
	}

	return json.Marshal(r.matches[0])
}
