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
	"sync"

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
	Count() int // Deprecated: Use MatchCount() for match count or Size() for element count
	MatchCount() int
	Keys() []string

	// Value access methods
	Error() error
	Value() (interface{}, error)
	Values() []interface{}
	Size() (int, error)

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
	mu sync.RWMutex
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
func (doc *Document) Query(xpath string) *Result {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.err != nil {
		return &Result{err: doc.err}
	}

	// Always materialize for now to ensure consistent behavior.
	// Future optimizations can re-introduce lazy parsing for specific simple paths.
	if !doc.isMaterialized {
		if err := doc.materialize(); err != nil {
			return &Result{err: err}
		}
	}

	// Handle special root cases
	if xpath == "" || xpath == "/" || xpath == "$" {
		return &Result{doc: doc, matches: []interface{}{doc.materialized}}
	}

	// Delegate all parsing to the parser
	p := parser.NewParser(xpath)
	query, err := p.Parse()
	if err != nil {
		// Return empty result for invalid paths instead of an error
		return &Result{doc: doc, matches: []interface{}{}}
	}

	// Use the engine to execute the parsed query
	matches, queryErr := engine.NewEngine().ExecuteOnMaterialized(doc.materialized, query)
	if queryErr != nil {
		return &Result{doc: doc, err: queryErr}
	}

	return &Result{
		doc:     doc,
		matches: convertMatches(matches),
	}
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
	d.mu.Lock()
	defer d.mu.Unlock()
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

	p := parser.NewParser(path)
	query, err := p.Parse()
	if err != nil {
		return ErrInvalidPath
	}

	// Use modifier to set the value
	return d.mod.Set(&d.materialized, query, value)
}

// Delete removes a value at the specified path, triggering materialization if needed
func (d *Document) Delete(path string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
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

	p := parser.NewParser(path)
	query, err := p.Parse()
	if err != nil {
		return ErrInvalidPath
	}

	// Use modifier to delete the value
	return d.mod.Delete(&d.materialized, query)
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
		// Only parse explicit boolean strings "true" and "false"
		if v == "true" {
			return true, nil
		}
		return false, ErrTypeMismatch
	case float64:
		// Only non-zero numbers are truthy, zero is falsy
		return v != 0, nil
	case int:
		// Only non-zero numbers are truthy, zero is falsy
		return v != 0, nil
	case int64:
		// Only non-zero numbers are truthy, zero is falsy
		return v != 0, nil
	default:
		return false, ErrTypeMismatch
	}
}

func (r *Result) MustBool() bool {
	b, err := r.Bool()
	if err != nil {
		panic(err)
	}
	return b
}

func (r *Result) Value() (interface{}, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.matches) == 0 {
		return nil, ErrNotFound
	}
	return r.matches[0], nil
}

func (r *Result) Values() []interface{} {
	if r.err != nil {
		return nil
	}
	return r.matches
}

func (r *Result) Error() error {
	return r.err
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
	// 统一用 parser 解析 path
	p := parser.NewParser(path)
	query, err := p.Parse()
	if err != nil {
		return &Result{err: fmt.Errorf("invalid sub-path syntax: %w", err)}
	}
	matches, queryErr := engine.NewEngine().ExecuteOnMaterialized(firstMatch, query)
	if queryErr != nil {
		return &Result{err: queryErr}
	}
	return &Result{
		doc:     r.doc,
		matches: convertMatches(matches),
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

// Count returns the count of matches or elements
func (r *Result) Count() int {
	if r.err != nil {
		return 0
	}

	if len(r.matches) == 0 {
		return 0
	}

	// If we have exactly one match, count based on its type
	if len(r.matches) == 1 {
		match := r.matches[0]
		if match == nil {
			return 0
		}
		if arr, ok := match.([]interface{}); ok {
			return len(arr)
		}
		if obj, ok := match.(map[string]interface{}); ok {
			return len(obj)
		}
		// For any other single primitive value, count is 1
		return 1
	}

	// Otherwise return the number of matches
	return len(r.matches)
}

func (r *Result) MatchCount() int {
	return len(r.matches)
}

func (r *Result) Size() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(r.matches) == 0 {
		return 0, ErrNotFound
	}
	match := r.matches[0]
	if arr, ok := match.([]interface{}); ok {
		return len(arr), nil
	}
	if obj, ok := match.(map[string]interface{}); ok {
		return len(obj), nil
	}
	return 0, ErrTypeMismatch
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

	// If a single match is an array, filter its elements and return a new result containing the filtered array.
	if len(r.matches) == 1 {
		if arr, ok := r.matches[0].([]interface{}); ok {
			var filteredArray []interface{} // This will be the new array
			for i, item := range arr {
				result := &Result{doc: r.doc, matches: []interface{}{item}}
				if fn(i, result) {
					filteredArray = append(filteredArray, item)
				}
			}
			// The result should contain the new filtered array as its single match
			return &Result{doc: r.doc, matches: []interface{}{filteredArray}}
		}
	}

	// Default behavior: filter the matches themselves. This is for when the result is already a list of matches.
	var filteredMatches []interface{}
	for i, match := range r.matches {
		result := &Result{doc: r.doc, matches: []interface{}{match}}
		if fn(i, result) {
			filteredMatches = append(filteredMatches, match)
		}
	}

	return &Result{
		doc:     r.doc,
		matches: filteredMatches,
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
	if r.err != nil {
		return false
	}

	// A result is considered an array if it has multiple matches,
	// OR if it has one match that is itself an array.
	if len(r.matches) > 1 {
		return true
	}

	if len(r.matches) == 1 {
		_, ok := r.matches[0].([]interface{})
		return ok
	}

	return false
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
	if len(r.matches) == 1 {
		return r.matches[0]
	}
	return r.matches
}

func (r *Result) Bytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.matches) == 0 {
		return nil, ErrNotFound
	}

	if len(r.matches) == 1 {
		return json.Marshal(r.matches[0])
	}
	return json.Marshal(r.matches)
}
