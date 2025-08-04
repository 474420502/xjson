package modifier

import (
	"errors"
	"strconv"
	"strings"
)

// Modifier handles path-based modifications on materialized JSON data
type Modifier struct{}

// NewModifier creates a new modifier instance
func NewModifier() *Modifier {
	return &Modifier{}
}

// Set sets a value at the specified path in the materialized data
func (m *Modifier) Set(data *interface{}, path string, value interface{}) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}

	if path == "" {
		*data = value
		return nil
	}

	// Parse the path into segments
	segments := m.parsePath(path)
	if len(segments) == 0 {
		return errors.New("invalid path")
	}

	// Navigate to the parent and set the final value
	return m.setAtPath(data, segments, value)
}

// Delete removes a value at the specified path in the materialized data
func (m *Modifier) Delete(data *interface{}, path string) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}

	if path == "" {
		return errors.New("cannot delete root")
	}

	// Parse the path into segments
	segments := m.parsePath(path)
	if len(segments) == 0 {
		return errors.New("invalid path")
	}

	// Navigate to the parent and delete the final key
	return m.deleteAtPath(data, segments)
}

// parsePath converts a dot-notation path into segments
func (m *Modifier) parsePath(path string) []string {
	if path == "" {
		return nil
	}

	// Simple implementation - split by dots
	// TODO: Handle escaped dots and array indices
	parts := strings.Split(path, ".")
	var segments []string

	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}

	return segments
}

// setAtPath navigates to the specified path and sets the value
func (m *Modifier) setAtPath(data *interface{}, segments []string, value interface{}) error {
	if len(segments) == 0 {
		*data = value
		return nil
	}

	current := *data

	// Ensure we have a container to work with
	if current == nil {
		current = make(map[string]interface{})
		*data = current
	}

	// Navigate to the parent of the target
	for i, segment := range segments[:len(segments)-1] {
		switch v := current.(type) {
		case map[string]interface{}:
			next, exists := v[segment]
			if !exists {
				// Create intermediate objects
				next = make(map[string]interface{})
				v[segment] = next
			}
			current = next
		case []interface{}:
			// Handle array index
			idx, err := strconv.Atoi(segment)
			if err != nil {
				return errors.New("invalid array index: " + segment)
			}
			if idx < 0 || idx >= len(v) {
				return errors.New("array index out of bounds: " + segment)
			}
			current = v[idx]
		default:
			return errors.New("cannot navigate through non-container type at segment: " + segments[i])
		}
	}

	// Set the final value
	finalSegment := segments[len(segments)-1]
	switch v := current.(type) {
	case map[string]interface{}:
		v[finalSegment] = value
	case []interface{}:
		idx, err := strconv.Atoi(finalSegment)
		if err != nil {
			return errors.New("invalid array index: " + finalSegment)
		}
		if idx < 0 || idx >= len(v) {
			return errors.New("array index out of bounds: " + finalSegment)
		}
		v[idx] = value
	default:
		return errors.New("cannot set value on non-container type")
	}

	return nil
}

// deleteAtPath navigates to the specified path and deletes the value
func (m *Modifier) deleteAtPath(data *interface{}, segments []string) error {
	if len(segments) == 0 {
		return errors.New("cannot delete root")
	}

	current := *data
	if current == nil {
		return errors.New("path not found")
	}

	// Navigate to the parent of the target
	for i, segment := range segments[:len(segments)-1] {
		switch v := current.(type) {
		case map[string]interface{}:
			next, exists := v[segment]
			if !exists {
				return errors.New("path not found at segment: " + segment)
			}
			current = next
		case []interface{}:
			idx, err := strconv.Atoi(segment)
			if err != nil {
				return errors.New("invalid array index: " + segment)
			}
			if idx < 0 || idx >= len(v) {
				return errors.New("array index out of bounds: " + segment)
			}
			current = v[idx]
		default:
			return errors.New("cannot navigate through non-container type at segment: " + segments[i])
		}
	}

	// Delete the final value
	finalSegment := segments[len(segments)-1]
	switch v := current.(type) {
	case map[string]interface{}:
		delete(v, finalSegment)
	case []interface{}:
		return errors.New("array element deletion not implemented yet")
	default:
		return errors.New("cannot delete from non-container type")
	}

	return nil
}
