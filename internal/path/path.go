// Package path provides utilities for working with JSON paths.
package path

import "strings"

// GetValueBySimplePath retrieves a value from a materialized structure using a simple path.
func GetValueBySimplePath(data interface{}, path string) (interface{}, bool) {
	// Path separator can be '/' for consistency.
	parts := strings.Split(strings.ReplaceAll(path, ".", "/"), "/")
	current := data
	for _, part := range parts {
		if part == "" {
			continue
		}
		if obj, ok := current.(map[string]interface{}); ok {
			if val, exists := obj[part]; exists {
				current = val
			} else {
				return nil, false
			}
		} else {
			// Cannot traverse further if not an object
			return nil, false
		}
	}
	return current, true
}
