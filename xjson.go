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
func (doc *Document) Query(path string) *Result {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.err != nil {
		return &Result{err: doc.err}
	}

	// Check for invalid syntax first, always
	if doc.isInvalidSyntax(path) {
		return &Result{matches: []interface{}{}}
	}

	// Handle empty path
	if path == "" {
		// Empty path should return the entire document (root)
		if !doc.isMaterialized {
			if err := doc.materialize(); err != nil {
				return &Result{err: err}
			}
		}
		return &Result{matches: []interface{}{doc.materialized}}
	}

	// Handle root path access
	if path == "/" {
		// Return the entire document
		if !doc.isMaterialized {
			if err := doc.materialize(); err != nil {
				return &Result{err: err}
			}
		}
		return &Result{matches: []interface{}{doc.materialized}}
	}

	// For read operations, use lazy parsing when possible
	if !doc.isMaterialized {

		// Try simple direct access first without materializing
		// Skip if it's a complex path (contains array access, recursive queries, or nested paths)
		isSimplePath := !strings.Contains(path, "[") &&
			!strings.Contains(path, "..") &&
			!strings.Contains(path, "//") &&
			!strings.Contains(path, ".") &&
			!(strings.HasPrefix(path, "/") && strings.Contains(path[1:], "/"))

		if isSimplePath {
			// Simple field access - try direct JSON parsing
			var data map[string]interface{}
			if err := json.Unmarshal(doc.raw, &data); err == nil {
				if val, exists := data[path]; exists {
					return &Result{matches: []interface{}{val}}
				}
			}
		}

		// For complex paths (both dot notation and XPath style), try path resolution
		if strings.Contains(path, ".") || strings.HasPrefix(path, "/") {
			var data interface{}
			if err := json.Unmarshal(doc.raw, &data); err == nil {
				current := data
				var parts []string
				if strings.HasPrefix(path, "/") {
					// XPath-style path: /user/profile/name or /user/orders[0]/total
					pathWithoutLeadingSlash := path[1:] // Remove leading slash
					parts = strings.Split(pathWithoutLeadingSlash, "/")
				} else {
					// Traditional dot notation: user.profile.name
					parts = strings.Split(path, ".")
				}

				found := true
				for _, part := range parts {
					if part == "" {
						continue
					}

					// Handle mixed field and array access like "orders[0]"
					if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
						// Split field name and array index
						bracketIndex := strings.Index(part, "[")
						fieldName := part[:bracketIndex]
						arrayPart := part[bracketIndex:]

						// First access the field
						if obj, ok := current.(map[string]interface{}); ok {
							if val, exists := obj[fieldName]; exists {
								current = val
							} else {
								found = false
								break
							}
						} else {
							found = false
							break
						}

						// Then handle array access
						indexStr := arrayPart[1 : len(arrayPart)-1]
						if index, err := strconv.Atoi(indexStr); err == nil {
							if arr, ok := current.([]interface{}); ok {
								if index >= 0 && index < len(arr) {
									current = arr[index]
									continue
								}
							}
						}
						found = false
						break
					}

					// Handle object field access
					if obj, ok := current.(map[string]interface{}); ok {
						if val, exists := obj[part]; exists {
							current = val
						} else {
							found = false
							break
						}
					} else {
						found = false
						break
					}
				}
				if found {
					return &Result{matches: []interface{}{current}}
				}
			}
		}

		// For complex queries, we need to materialize
		if err := doc.materialize(); err != nil {
			return &Result{err: err}
		}
	}

	// We need to check if the key exists, not just if the value is non-nil
	exists, result := doc.getValueWithExists(doc.materialized, path)
	if !exists {
		return &Result{matches: []interface{}{}}
	}

	return &Result{matches: []interface{}{result}}
}

// getValueWithExists checks if a path exists and returns the value
func (doc *Document) getValueWithExists(data interface{}, path string) (bool, interface{}) {
	if path == "" {
		return true, data
	}

	// Handle direct key access first (including keys with dots)
	if obj, ok := data.(map[string]interface{}); ok {
		if val, exists := obj[path]; exists {
			return true, val
		}
	}

	// Handle recursive queries (..) or (//) early before other path processing
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return doc.handleRecursiveQuery(data, path)
	}

	// Handle simple paths with array access or filters (like "products[0]" or "products[?(@.field == value)]")
	if strings.Contains(path, "[") && strings.HasSuffix(path, "]") && !strings.Contains(path, ".") && !strings.HasPrefix(path, "/") {
		// Split field name and array/filter part
		bracketIndex := strings.Index(path, "[")
		fieldName := path[:bracketIndex]
		arrayPart := path[bracketIndex:]

		// Handle the case where fieldName is empty (pure array access like "[0]")
		var current interface{} = data
		if fieldName != "" {
			// First access the field
			if obj, ok := data.(map[string]interface{}); ok {
				if val, exists := obj[fieldName]; exists {
					current = val
				} else {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
		// If fieldName is empty, current is already data (root level array access)

		// Then handle the array access or filter
		if strings.HasPrefix(arrayPart, "[") && strings.HasSuffix(arrayPart, "]") {
			indexStr := arrayPart[1 : len(arrayPart)-1]

			// Handle filter expressions like [?(@.field == value)]
			if strings.HasPrefix(indexStr, "?(") && strings.HasSuffix(indexStr, ")") {
				if arr, ok := current.([]interface{}); ok {
					filtered := doc.applyFilter(arr, indexStr)
					return len(filtered) > 0, filtered
				}
				return false, nil
			}

			// Handle slice syntax [start:end]
			if strings.Contains(indexStr, ":") {
				parts := strings.Split(indexStr, ":")
				if len(parts) == 2 {
					var start, end int
					var err1, err2 error

					// Handle open-ended slices like [1:] and [:2]
					if parts[0] == "" {
						start = 0 // [:end] starts from 0
					} else {
						start, err1 = strconv.Atoi(parts[0])
					}

					if parts[1] == "" {
						// [start:] goes to the end - we'll set end later
						end = -1 // Use -1 as a marker for "to the end"
						err2 = nil
					} else {
						end, err2 = strconv.Atoi(parts[1])
					}

					if err1 == nil && err2 == nil {
						if arr, ok := current.([]interface{}); ok {
							// Handle negative indices for slice
							if start < 0 {
								start = len(arr) + start
							}
							if end == -1 {
								end = len(arr) // [start:] goes to the end
							} else if end < 0 {
								end = len(arr) + end
							}
							// Clamp to array bounds
							if start < 0 {
								start = 0
							}
							if end > len(arr) {
								end = len(arr)
							}
							if start <= end && start < len(arr) {
								slice := arr[start:end]
								return len(slice) > 0, slice
							}
						}
					}
				}
				return false, nil
			}

			// Handle single index
			if index, err := strconv.Atoi(indexStr); err == nil {
				if arr, ok := current.([]interface{}); ok {
					// Handle negative indices
					if index < 0 {
						index = len(arr) + index
					}
					if index >= 0 && index < len(arr) {
						return true, arr[index]
					}
				}
			}
			return false, nil
		}
	}

	// Handle XPath-style or dotted paths
	if strings.Contains(path, ".") || strings.HasPrefix(path, "/") {
		var parts []string
		if strings.HasPrefix(path, "/") {
			// XPath-style path: /user/profile/name or /user/orders[0]/total
			pathWithoutLeadingSlash := path[1:] // Remove leading slash
			parts = strings.Split(pathWithoutLeadingSlash, "/")
		} else {
			// Traditional dot notation: user.profile.name
			parts = doc.splitPath(path)
		}

		current := data
		for _, part := range parts {
			if part == "" {
				continue
			}

			// Handle mixed field and array access like "orders[0]"
			if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
				// Split field name and array index
				bracketIndex := strings.Index(part, "[")
				fieldName := part[:bracketIndex]
				arrayPart := part[bracketIndex:]

				// First access the field
				if obj, ok := current.(map[string]interface{}); ok {
					if val, exists := obj[fieldName]; exists {
						current = val
					} else {
						return false, nil
					}
				} else {
					return false, nil
				}

				// Then handle array access
				indexStr := arrayPart[1 : len(arrayPart)-1]

				// Handle filter expressions like [?(@.field == value)]
				if strings.HasPrefix(indexStr, "?(") && strings.HasSuffix(indexStr, ")") {
					if arr, ok := current.([]interface{}); ok {
						filtered := doc.applyFilter(arr, indexStr)
						current = filtered
						continue
					}
					return false, nil
				}

				// Handle slice syntax [start:end]
				if strings.Contains(indexStr, ":") {
					parts := strings.Split(indexStr, ":")
					if len(parts) == 2 {
						var start, end int
						var err1, err2 error

						// Handle open-ended slices like [1:] and [:2]
						if parts[0] == "" {
							start = 0 // [:end] starts from 0
						} else {
							start, err1 = strconv.Atoi(parts[0])
						}

						if parts[1] == "" {
							// [start:] goes to the end - we'll set end later
							end = -1 // Use -1 as a marker for "to the end"
							err2 = nil
						} else {
							end, err2 = strconv.Atoi(parts[1])
						}

						if err1 == nil && err2 == nil {
							if arr, ok := current.([]interface{}); ok {
								// Handle negative indices for slice
								if start < 0 {
									start = len(arr) + start
								}
								if end == -1 {
									end = len(arr) // [start:] goes to the end
								} else if end < 0 {
									end = len(arr) + end
								}
								// Clamp to array bounds
								if start < 0 {
									start = 0
								}
								if end > len(arr) {
									end = len(arr)
								}
								if start <= end && start < len(arr) {
									slice := arr[start:end]
									current = slice
									continue
								}
							}
						}
					}
					return false, nil
				}

				// Handle single index
				if index, err := strconv.Atoi(indexStr); err == nil {
					if arr, ok := current.([]interface{}); ok {
						// Handle negative indices
						if index < 0 {
							index = len(arr) + index
						}
						if index >= 0 && index < len(arr) {
							current = arr[index]
							continue
						}
					}
				}
				return false, nil
			}

			// Handle pure array access like "[0]"
			if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
				indexStr := part[1 : len(part)-1]
				if index, err := strconv.Atoi(indexStr); err == nil {
					if arr, ok := current.([]interface{}); ok {
						if index >= 0 && index < len(arr) {
							current = arr[index]
							continue
						}
					}
				}
				return false, nil
			}

			// Handle object field access
			if obj, ok := current.(map[string]interface{}); ok {
				if val, exists := obj[part]; exists {
					current = val
				} else {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
		return true, current
	}

	// Handle simple field access (no dots, no brackets, no XPath)
	if !strings.Contains(path, ".") && !strings.Contains(path, "[") && !strings.HasPrefix(path, "/") {
		return false, nil // Already checked above
	}

	// Handle array access at root level like [0] or [1:3]
	if strings.HasPrefix(path, "[") && strings.HasSuffix(path, "]") {
		indexStr := path[1 : len(path)-1]

		// Handle slice syntax [start:end]
		if strings.Contains(indexStr, ":") {
			parts := strings.Split(indexStr, ":")
			if len(parts) == 2 {
				var start, end int
				var err1, err2 error

				// Handle open-ended slices like [1:] and [:2]
				if parts[0] == "" {
					start = 0 // [:end] starts from 0
				} else {
					start, err1 = strconv.Atoi(parts[0])
				}

				if parts[1] == "" {
					// [start:] goes to the end - we'll set end later
					end = -1 // Use -1 as a marker for "to the end"
					err2 = nil
				} else {
					end, err2 = strconv.Atoi(parts[1])
				}

				if err1 == nil && err2 == nil {
					if arr, ok := data.([]interface{}); ok {
						// Handle negative indices for slice
						if start < 0 {
							start = len(arr) + start
						}
						if end == -1 {
							end = len(arr) // [start:] goes to the end
						} else if end < 0 {
							end = len(arr) + end
						}
						// Clamp to array bounds
						if start < 0 {
							start = 0
						}
						if end > len(arr) {
							end = len(arr)
						}
						if start <= end && start < len(arr) {
							slice := arr[start:end]
							return true, slice
						}
					}
				}
			}
			return false, nil
		}

		// Handle negative indices
		if strings.HasPrefix(indexStr, "-") {
			if index, err := strconv.Atoi(indexStr); err == nil {
				if arr, ok := data.([]interface{}); ok {
					if index < 0 {
						index = len(arr) + index
					}
					if index >= 0 && index < len(arr) {
						return true, arr[index]
					}
				}
			}
			return false, nil
		}

		// Handle positive indices
		if index, err := strconv.Atoi(indexStr); err == nil {
			if arr, ok := data.([]interface{}); ok {
				if index >= 0 && index < len(arr) {
					return true, arr[index]
				}
			}
		}
		return false, nil
	}

	// Handle dotted paths by splitting on dots (but not inside brackets)
	parts := doc.splitPath(path)
	current := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle array access like [0] or [-1]
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			indexStr := part[1 : len(part)-1]

			var index int
			var err error

			// Handle negative indices
			if strings.HasPrefix(indexStr, "-") {
				index, err = strconv.Atoi(indexStr)
				if err != nil {
					return false, nil
				}
				if arr, ok := current.([]interface{}); ok {
					if index < 0 {
						index = len(arr) + index
					}
					if index >= 0 && index < len(arr) {
						current = arr[index]
						continue
					}
				}
				return false, nil
			} else {
				// Handle positive indices
				index, err = strconv.Atoi(indexStr)
				if err != nil {
					return false, nil
				}
				if arr, ok := current.([]interface{}); ok {
					if index >= 0 && index < len(arr) {
						current = arr[index]
						continue
					}
				}
				return false, nil
			}
		}

		// Handle combined field and array access like "book[0]" or "book[1:3]" or "products[?(@.field == value)]"
		if strings.Contains(part, "[") {
			// Split field and array access
			bracketIndex := strings.Index(part, "[")
			fieldName := part[:bracketIndex]
			arrayPart := part[bracketIndex:]

			// First access the field
			if obj, ok := current.(map[string]interface{}); ok {
				if val, exists := obj[fieldName]; exists {
					current = val
				} else {
					return false, nil
				}
			} else {
				return false, nil
			}

			// Then handle the array access or filter
			if strings.HasPrefix(arrayPart, "[") && strings.HasSuffix(arrayPart, "]") {
				indexStr := arrayPart[1 : len(arrayPart)-1]

				// Handle filter expressions like [?(@.field == value)]
				if strings.HasPrefix(indexStr, "?(") && strings.HasSuffix(indexStr, ")") {
					if arr, ok := current.([]interface{}); ok {
						filtered := doc.applyFilter(arr, indexStr)
						current = filtered
						continue
					}
					return false, nil
				}

				// Handle slice syntax [start:end]
				if strings.Contains(indexStr, ":") {
					parts := strings.Split(indexStr, ":")
					if len(parts) == 2 {
						var start, end int
						var err1, err2 error

						// Handle open-ended slices like [1:] and [:2]
						if parts[0] == "" {
							start = 0 // [:end] starts from 0
						} else {
							start, err1 = strconv.Atoi(parts[0])
						}

						if parts[1] == "" {
							// [start:] goes to the end - we'll set end later
							end = -1 // Use -1 as a marker for "to the end"
							err2 = nil
						} else {
							end, err2 = strconv.Atoi(parts[1])
						}

						if err1 == nil && err2 == nil {
							if arr, ok := current.([]interface{}); ok {
								// Handle negative indices for slice
								if start < 0 {
									start = len(arr) + start
								}
								if end == -1 {
									end = len(arr) // [start:] goes to the end
								} else if end < 0 {
									end = len(arr) + end
								}
								// Clamp to array bounds
								if start < 0 {
									start = 0
								}
								if end > len(arr) {
									end = len(arr)
								}
								if start <= end && start < len(arr) {
									slice := arr[start:end]
									current = slice
									continue
								}
							}
						}
					}
					return false, nil
				}

				var index int
				var err error

				// Handle negative indices
				if strings.HasPrefix(indexStr, "-") {
					index, err = strconv.Atoi(indexStr)
					if err != nil {
						return false, nil
					}
					if arr, ok := current.([]interface{}); ok {
						if index < 0 {
							index = len(arr) + index
						}
						if index >= 0 && index < len(arr) {
							current = arr[index]
							continue
						}
					}
					return false, nil
				} else {
					// Handle positive indices
					index, err = strconv.Atoi(indexStr)
					if err != nil {
						return false, nil
					}
					if arr, ok := current.([]interface{}); ok {
						if index >= 0 && index < len(arr) {
							current = arr[index]
							continue
						}
					}
					return false, nil
				}
			}
			continue
		}

		// Handle object field access
		if obj, ok := current.(map[string]interface{}); ok {
			if val, exists := obj[part]; exists {
				current = val
			} else {
				return false, nil
			}
		} else {
			return false, nil
		}
	}

	return true, current
}

// handleRecursiveQuery handles queries with ".." or "//" syntax for recursive search
func (doc *Document) handleRecursiveQuery(data interface{}, path string) (bool, interface{}) {
	// Parse the recursive query
	// For example: "store..price" or "store//price" means find "price" in store and all its descendants
	// or "..price" or "//price" means find all "price" fields at any depth

	if strings.HasPrefix(path, "//") {
		// Simple recursive query from root (//fieldName or //fieldName[filter])
		fieldPart := strings.TrimPrefix(path, "//")

		// Check if there's a filter expression
		if strings.Contains(fieldPart, "[") {
			// Parse field name and filter
			bracketIndex := strings.Index(fieldPart, "[")
			fieldName := fieldPart[:bracketIndex]
			filterPart := fieldPart[bracketIndex:]

			// Find all instances of the field recursively
			fieldResults := doc.findAllFields(data, fieldName, 0)
			var filteredResults []interface{}

			// Apply filter to each found array
			for _, fieldResult := range fieldResults {
				if arr, ok := fieldResult.([]interface{}); ok {
					// Apply the filter expression to this array
					if strings.HasPrefix(filterPart, "[") && strings.HasSuffix(filterPart, "]") {
						indexStr := filterPart[1 : len(filterPart)-1]
						if strings.HasPrefix(indexStr, "?(") && strings.HasSuffix(indexStr, ")") {
							filtered := doc.applyFilter(arr, indexStr)
							filteredResults = append(filteredResults, filtered...)
						}
					}
				}
			}

			if len(filteredResults) > 0 {
				return true, filteredResults
			}
			return false, nil
		} else {
			// Simple field name without filter
			results := doc.findAllFields(data, fieldPart, 0)
			if len(results) > 0 {
				// If we found multiple results, return them as an array
				if len(results) == 1 {
					return true, results[0]
				}
				return true, results
			}
			return false, nil
		}
	}

	if strings.HasPrefix(path, "..") {
		// Simple recursive query from root (..fieldName)
		fieldName := strings.TrimPrefix(path, "..")
		results := doc.findAllFields(data, fieldName, 0)
		if len(results) > 0 {
			// If we found multiple results, return them as an array
			if len(results) == 1 {
				return true, results[0]
			}
			return true, results
		}
		return false, nil
	}

	// Handle complex recursive queries like "store//price" or "store..price"
	var parts []string
	var fieldName string
	var prefixPath string

	if strings.Contains(path, "//") {
		parts = strings.Split(path, "//")
		if len(parts) == 2 {
			prefixPath = parts[0]
			fieldName = parts[1]
		}
	} else if strings.Contains(path, "..") {
		parts = strings.Split(path, "..")
		if len(parts) == 2 {
			prefixPath = parts[0]
			fieldName = parts[1]
		}
	}

	if prefixPath != "" && fieldName != "" {
		// First navigate to the prefix path
		exists, prefixData := doc.getValueWithExists(data, prefixPath)
		if !exists {
			return false, nil
		}

		// Then do recursive search from that point
		results := doc.findAllFields(prefixData, fieldName, 0)
		if len(results) > 0 {
			// If we found multiple results, return them as an array
			if len(results) == 1 {
				return true, results[0]
			}
			return true, results
		}
		return false, nil
	}

	return false, nil
}

// findAllFields recursively searches for all fields with the given name
func (doc *Document) findAllFields(data interface{}, fieldName string, depth int) []interface{} {
	var results []interface{}

	// Prevent infinite recursion
	if depth > 50 {
		return results
	}

	switch v := data.(type) {
	case map[string]interface{}:
		// Check if current object has the target field
		if value, exists := v[fieldName]; exists {
			results = append(results, value)
		}

		// Recursively search all child objects
		for _, child := range v {
			subResults := doc.findAllFields(child, fieldName, depth+1)
			results = append(results, subResults...)
		}
	case []interface{}:
		// Recursively search array elements
		for _, child := range v {
			subResults := doc.findAllFields(child, fieldName, depth+1)
			results = append(results, subResults...)
		}
	}

	return results
}

// getValue extracts a value from materialized data using a simple path
func (doc *Document) getValue(data interface{}, path string) interface{} {
	if path == "" {
		return data
	}

	// Handle direct key access first (including keys with dots)
	if obj, ok := data.(map[string]interface{}); ok {
		if val, exists := obj[path]; exists {
			// 重要：即使值为nil，也要返回一个特殊标记表示"找到了"
			// 我们需要区分"键不存在"和"键存在但值为null"
			return val
		}
	}

	// Handle simple field access (no dots, no brackets)
	if !strings.Contains(path, ".") && !strings.Contains(path, "[") {
		return nil // Already checked above
	}

	// Handle array access at root level like [0]
	if strings.HasPrefix(path, "[") && strings.HasSuffix(path, "]") {
		indexStr := path[1 : len(path)-1]

		// Handle negative indices
		if strings.HasPrefix(indexStr, "-") {
			if index, err := strconv.Atoi(indexStr); err == nil {
				if arr, ok := data.([]interface{}); ok {
					if index < 0 {
						index = len(arr) + index
					}
					if index >= 0 && index < len(arr) {
						return arr[index]
					}
				}
			}
			return nil
		}

		// Handle positive indices
		if index, err := strconv.Atoi(indexStr); err == nil {
			if arr, ok := data.([]interface{}); ok {
				if index >= 0 && index < len(arr) {
					return arr[index]
				}
			}
		}
		return nil
	}

	// Handle dotted paths or XPath-style paths
	var parts []string
	if strings.HasPrefix(path, "/") {
		// XPath-style path: /user/profile/name
		parts = strings.Split(path, "/")
		// Remove empty first element from leading slash
		if len(parts) > 0 && parts[0] == "" {
			parts = parts[1:]
		}
	} else {
		// Traditional dot notation: user.profile.name
		parts = strings.Split(path, ".")
	}
	current := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle array access like [0] or [-1]
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			indexStr := part[1 : len(part)-1]

			var index int
			var err error

			// Handle negative indices
			if strings.HasPrefix(indexStr, "-") {
				index, err = strconv.Atoi(indexStr)
				if err != nil {
					return nil
				}
				if arr, ok := current.([]interface{}); ok {
					if index < 0 {
						index = len(arr) + index
					}
					if index >= 0 && index < len(arr) {
						current = arr[index]
						continue
					}
				}
				return nil
			} else {
				// Handle positive indices
				index, err = strconv.Atoi(indexStr)
				if err != nil {
					return nil
				}
				if arr, ok := current.([]interface{}); ok {
					if index >= 0 && index < len(arr) {
						current = arr[index]
						continue
					}
				}
				return nil
			}
		}

		// Handle combined field and array access like "book[0]"
		if strings.Contains(part, "[") {
			// Split field and array access
			bracketIndex := strings.Index(part, "[")
			fieldName := part[:bracketIndex]
			arrayPart := part[bracketIndex:]

			// First access the field
			if obj, ok := current.(map[string]interface{}); ok {
				current = obj[fieldName]
			} else {
				return nil
			}

			// Then handle the array access
			if strings.HasPrefix(arrayPart, "[") && strings.HasSuffix(arrayPart, "]") {
				indexStr := arrayPart[1 : len(arrayPart)-1]

				var index int
				var err error

				// Handle negative indices
				if strings.HasPrefix(indexStr, "-") {
					index, err = strconv.Atoi(indexStr)
					if err != nil {
						return nil
					}
					if arr, ok := current.([]interface{}); ok {
						if index < 0 {
							index = len(arr) + index
						}
						if index >= 0 && index < len(arr) {
							current = arr[index]
							continue
						}
					}
					return nil
				} else {
					// Handle positive indices
					index, err = strconv.Atoi(indexStr)
					if err != nil {
						return nil
					}
					if arr, ok := current.([]interface{}); ok {
						if index >= 0 && index < len(arr) {
							current = arr[index]
							continue
						}
					}
					return nil
				}
			}
			continue
		}

		// Handle object field access
		if obj, ok := current.(map[string]interface{}); ok {
			current = obj[part]
		} else {
			return nil
		}
	}

	return current
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

	// Use modifier to set the value
	return d.mod.Set(&d.materialized, path, value)
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
		// Only parse explicit boolean strings "true" and "false"
		if v == "true" {
			return true, nil
		}
		if v == "false" {
			return false, nil
		}
		// Any other string should return type mismatch error
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
	case map[string]interface{}:
		// Objects are truthy (non-nil)
		return true, nil
	case []interface{}:
		// Arrays are truthy (non-empty)
		return true, nil
	default:
		// For any other type, return type mismatch error
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

	// If we have exactly one match
	if len(r.matches) == 1 {
		value := r.matches[0]
		if value == nil {
			return 1 // null values have count 1 (they exist but are null)
		}
		if arr, ok := value.([]interface{}); ok {
			return len(arr) // arrays return their length
		}
		if obj, ok := value.(map[string]interface{}); ok {
			return len(obj) // objects return their field count
		}
		// For other single values, return 1 (the match count)
		return 1
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
	if r.err != nil {
		return false
	}

	// If we have multiple matches, this is an array result
	if len(r.matches) > 1 {
		return true
	}

	// If we have exactly one match, check if it's an array type
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
	// For null values, return nil (which is correct Go representation)
	value := r.matches[0]
	if value == nil {
		return nil
	}
	return value
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

// Error returns the error that occurred during the query, if any.
func (r *Result) Error() error {
	return r.err
}

// MatchCount returns the number of nodes that matched the query.
func (r *Result) MatchCount() int {
	if r.err != nil {
		return 0
	}
	return len(r.matches)
}

// Value returns the first matched value. It returns an error if there is not
// exactly one match. This is useful for queries that are expected to return a single node.
func (r *Result) Value() (interface{}, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.matches) != 1 {
		return nil, fmt.Errorf("expected exactly 1 match, but got %d", len(r.matches))
	}
	return r.matches[0], nil
}

// Values returns a slice of all matched values.
func (r *Result) Values() []interface{} {
	return r.matches
}

// Size returns the element count of the single matched node.
// It is a convenience method that only works if the query resulted in
// exactly one match which is an array or an object.
// For arrays, it returns the number of elements.
// For objects, it returns the number of key-value pairs.
// For all other types (including null), it returns 1.
// It returns 0 and an error if there was a query error or not exactly one match.
func (r *Result) Size() (int, error) {
	value, err := r.Value() // Reuse our new, robust Value() method
	if err != nil {
		return 0, err
	}

	if value == nil {
		return 1, nil // A single null value exists. Its size is 1.
	}
	if arr, ok := value.([]interface{}); ok {
		return len(arr), nil
	}
	if obj, ok := value.(map[string]interface{}); ok {
		return len(obj), nil
	}
	// Any other single, non-collection value has a size of 1.
	return 1, nil
}

// isInvalidSyntax checks for basic syntax errors in query paths
func (doc *Document) isInvalidSyntax(path string) bool {
	// Check for incomplete filter expressions like "a[?("
	if strings.Contains(path, "[?(") && !strings.Contains(path, ")]") {
		return true
	}

	// Check for unmatched brackets
	openBrackets := strings.Count(path, "[")
	closeBrackets := strings.Count(path, "]")
	if openBrackets != closeBrackets {
		return true
	}

	// Check for invalid slice syntax like [1:2:3]
	if strings.Contains(path, ":") {
		// Find bracket contents
		start := strings.Index(path, "[")
		end := strings.LastIndex(path, "]")
		if start != -1 && end != -1 && start < end {
			content := path[start+1 : end]
			if strings.Count(content, ":") > 1 {
				return true
			}
		}
	}

	return false
}

// applyFilter applies a filter expression to an array
func (doc *Document) applyFilter(arr []interface{}, filterExpr string) []interface{} {
	// Remove the leading "?(" and trailing ")"
	expr := strings.TrimPrefix(filterExpr, "?(")
	expr = strings.TrimSuffix(expr, ")")

	var filtered []interface{}

	for _, item := range arr {
		if doc.evaluateFilterExpression(item, expr) {
			filtered = append(filtered, item)
		}
	}

	return filtered
} // evaluateFilterExpression evaluates a filter expression against an item
func (doc *Document) evaluateFilterExpression(item interface{}, expr string) bool {
	// Handle AND operator first (higher precedence)
	if strings.Contains(expr, " && ") {
		parts := strings.Split(expr, " && ")
		for _, part := range parts {
			if !doc.evaluateSimpleExpression(item, strings.TrimSpace(part)) {
				return false // All parts must be true for AND
			}
		}
		return true
	}

	// Handle OR operator
	if strings.Contains(expr, " || ") {
		parts := strings.Split(expr, " || ")
		for _, part := range parts {
			if doc.evaluateSimpleExpression(item, strings.TrimSpace(part)) {
				return true // Any part can be true for OR
			}
		}
		return false
	}

	// Handle simple expression
	return doc.evaluateSimpleExpression(item, expr)
}

// evaluateSimpleExpression evaluates a simple filter expression against an item
func (doc *Document) evaluateSimpleExpression(item interface{}, expr string) bool {
	// Trim spaces
	expr = strings.TrimSpace(expr)

	// Must start with @.field (or @.field.sub)
	if !strings.HasPrefix(expr, "@.") {
		return false
	}

	// Supported operators
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}
	var op string
	var idx int
	for _, candidate := range operators {
		if i := strings.Index(expr, candidate); i > 0 {
			op = candidate
			idx = i
			break
		}
	}
	if op == "" {
		return false // no operator found
	}

	fieldPath := strings.TrimSpace(expr[:idx])
	expectedValue := strings.TrimSpace(expr[idx+len(op):])

	// Remove quotes from expectedValue if present
	if (strings.HasPrefix(expectedValue, "'") && strings.HasSuffix(expectedValue, "'")) ||
		(strings.HasPrefix(expectedValue, "\"") && strings.HasSuffix(expectedValue, "\"")) {
		expectedValue = expectedValue[1 : len(expectedValue)-1]
	}

	// Support nested field access: @.a.b.c or @/a/b/c
	fieldName := fieldPath[2:]
	var actualValue interface{} = item

	// Parse field path segments
	var parts []string
	if strings.HasPrefix(fieldName, "/") {
		// XPath-style path: /a/b/c
		parts = strings.Split(fieldName, "/")
		// Remove empty first element from leading slash
		if len(parts) > 0 && parts[0] == "" {
			parts = parts[1:]
		}
	} else {
		// Traditional dot notation: a.b.c
		parts = strings.Split(fieldName, ".")
	}

	for _, part := range parts {
		if m, ok := actualValue.(map[string]interface{}); ok {
			actualValue = m[part]
		} else {
			actualValue = nil
			break
		}
	}

	// If field doesn't exist, always false
	if actualValue == nil {
		if op == "==" && (expectedValue == "null" || expectedValue == "nil") {
			return true
		}
		return false
	}

	// Use compareValues for all logic
	return doc.compareValues(actualValue, expectedValue, op)
}

// compareValues compares two values based on the operator
func (doc *Document) compareValues(actual interface{}, expected string, operator string) bool {
	// nil/null handling
	isNil := func(v interface{}) bool {
		return v == nil
	}

	switch operator {
	case "==":
		// nil/null equality
		if isNil(actual) && (expected == "null" || expected == "nil") {
			return true
		}
		if (expected == "null" || expected == "nil") && !isNil(actual) {
			return false
		}
		if isNil(actual) && expected != "null" && expected != "nil" {
			return false
		}

		// Boolean comparison
		if expected == "true" || expected == "false" {
			expectedBool := expected == "true"
			switch v := actual.(type) {
			case bool:
				return v == expectedBool
			case int:
				return (v != 0) == expectedBool
			case float64:
				return (v != 0) == expectedBool
			case string:
				return (v == "true") == expectedBool
			default:
				return false
			}
		}

		// String comparison (remove quotes)
		if strings.HasPrefix(expected, "'") && strings.HasSuffix(expected, "'") {
			expectedStr := strings.Trim(expected, "'")
			if actualStr, ok := actual.(string); ok {
				return actualStr == expectedStr
			}
		}
		if strings.HasPrefix(expected, "\"") && strings.HasSuffix(expected, "\"") {
			expectedStr := strings.Trim(expected, "\"")
			if actualStr, ok := actual.(string); ok {
				return actualStr == expectedStr
			}
		}
		// String comparison without quotes
		if actualStr, ok := actual.(string); ok {
			return actualStr == expected
		}

		// Numeric comparison
		if expectedFloat, err := strconv.ParseFloat(expected, 64); err == nil {
			switch v := actual.(type) {
			case float64:
				return v == expectedFloat
			case int:
				return float64(v) == expectedFloat
			case int64:
				return float64(v) == expectedFloat
			case string:
				if actualFloat, err := strconv.ParseFloat(v, 64); err == nil {
					return actualFloat == expectedFloat
				}
			}
		}
		return false

	case "!=":
		// nil/null
		if isNil(actual) && (expected == "null" || expected == "nil") {
			return false
		}
		if (expected == "null" || expected == "nil") && !isNil(actual) {
			return true
		}
		if isNil(actual) && expected != "null" && expected != "nil" {
			return true
		}

		// Boolean
		if expected == "true" || expected == "false" {
			expectedBool := expected == "true"
			switch v := actual.(type) {
			case bool:
				return v != expectedBool
			case int:
				return (v != 0) != expectedBool
			case float64:
				return (v != 0) != expectedBool
			case string:
				return (v == "true") != expectedBool
			default:
				return true
			}
		}

		// String (remove quotes)
		if strings.HasPrefix(expected, "'") && strings.HasSuffix(expected, "'") {
			expectedStr := strings.Trim(expected, "'")
			if actualStr, ok := actual.(string); ok {
				return actualStr != expectedStr
			}
		}
		if strings.HasPrefix(expected, "\"") && strings.HasSuffix(expected, "\"") {
			expectedStr := strings.Trim(expected, "\"")
			if actualStr, ok := actual.(string); ok {
				return actualStr != expectedStr
			}
		}
		if actualStr, ok := actual.(string); ok {
			return actualStr != expected
		}

		// Numeric
		if expectedFloat, err := strconv.ParseFloat(expected, 64); err == nil {
			switch v := actual.(type) {
			case float64:
				return v != expectedFloat
			case int:
				return float64(v) != expectedFloat
			case int64:
				return float64(v) != expectedFloat
			case string:
				if actualFloat, err := strconv.ParseFloat(v, 64); err == nil {
					return actualFloat != expectedFloat
				}
			}
		}
		return true

	case "<":
		if expectedFloat, err := strconv.ParseFloat(expected, 64); err == nil {
			switch v := actual.(type) {
			case float64:
				return v < expectedFloat
			case int:
				return float64(v) < expectedFloat
			case int64:
				return float64(v) < expectedFloat
			case string:
				if actualFloat, err := strconv.ParseFloat(v, 64); err == nil {
					return actualFloat < expectedFloat
				}
			}
		}
		return false

	case ">":
		if expectedFloat, err := strconv.ParseFloat(expected, 64); err == nil {
			switch v := actual.(type) {
			case float64:
				return v > expectedFloat
			case int:
				return float64(v) > expectedFloat
			case int64:
				return float64(v) > expectedFloat
			case string:
				if actualFloat, err := strconv.ParseFloat(v, 64); err == nil {
					return actualFloat > expectedFloat
				}
			}
		}
		return false

	case "<=":
		if expectedFloat, err := strconv.ParseFloat(expected, 64); err == nil {
			switch v := actual.(type) {
			case float64:
				return v <= expectedFloat
			case int:
				return float64(v) <= expectedFloat
			case int64:
				return float64(v) <= expectedFloat
			case string:
				if actualFloat, err := strconv.ParseFloat(v, 64); err == nil {
					return actualFloat <= expectedFloat
				}
			}
		}
		return false

	case ">=":
		if expectedFloat, err := strconv.ParseFloat(expected, 64); err == nil {
			switch v := actual.(type) {
			case float64:
				return v >= expectedFloat
			case int:
				return float64(v) >= expectedFloat
			case int64:
				return float64(v) >= expectedFloat
			case string:
				if actualFloat, err := strconv.ParseFloat(v, 64); err == nil {
					return actualFloat >= expectedFloat
				}
			}
		}
		return false
	}

	return false
}

// splitPath splits a path on dots but ignores dots inside bracket expressions
func (doc *Document) splitPath(path string) []string {
	var parts []string
	var current strings.Builder
	bracketDepth := 0

	for _, char := range path {
		switch char {
		case '[':
			bracketDepth++
			current.WriteRune(char)
		case ']':
			bracketDepth--
			current.WriteRune(char)
		case '.':
			if bracketDepth == 0 {
				// We're not inside brackets, so this dot is a path separator
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				// We're inside brackets, so this dot is part of the expression
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
