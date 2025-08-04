package modifier

import (
	"reflect"
	"testing"
)

func TestNewModifier(t *testing.T) {
	m := NewModifier()
	if m == nil {
		t.Error("NewModifier() returned nil")
	}
}

func TestModifierSet(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		path        string
		value       interface{}
		expected    interface{}
		expectError bool
	}{
		{
			name:     "set root value",
			data:     map[string]interface{}{"a": 1},
			path:     "",
			value:    "new root",
			expected: "new root",
		},
		{
			name:     "set simple property",
			data:     map[string]interface{}{"name": "John"},
			path:     "name",
			value:    "Jane",
			expected: map[string]interface{}{"name": "Jane"},
		},
		{
			name:     "set nested property",
			data:     map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
			path:     "user.name",
			value:    "Jane",
			expected: map[string]interface{}{"user": map[string]interface{}{"name": "Jane"}},
		},
		{
			name:     "create new property",
			data:     map[string]interface{}{"name": "John"},
			path:     "age",
			value:    30,
			expected: map[string]interface{}{"name": "John", "age": 30},
		},
		{
			name:     "create nested property",
			data:     map[string]interface{}{},
			path:     "user.name",
			value:    "John",
			expected: map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
		},
		{
			name:     "set array element",
			data:     map[string]interface{}{"items": []interface{}{"a", "b", "c"}},
			path:     "items.1",
			value:    "new_b",
			expected: map[string]interface{}{"items": []interface{}{"a", "new_b", "c"}},
		},
		{
			name:     "set on nil data",
			data:     nil,
			path:     "name",
			value:    "John",
			expected: map[string]interface{}{"name": "John"},
		},
		{
			name:        "nil data pointer",
			data:        nil,
			path:        "test",
			value:       "value",
			expectError: true,
		},
		{
			name:        "invalid array index",
			data:        map[string]interface{}{"items": []interface{}{"a", "b"}},
			path:        "items.abc",
			value:       "test",
			expectError: true,
		},
		{
			name:        "array index out of bounds",
			data:        map[string]interface{}{"items": []interface{}{"a", "b"}},
			path:        "items.5",
			value:       "test",
			expectError: true,
		},
		{
			name:        "navigate through non-container",
			data:        map[string]interface{}{"value": "string"},
			path:        "value.subfield",
			value:       "test",
			expectError: true,
		},
		{
			name:        "set on non-container",
			data:        "string",
			path:        "field",
			value:       "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModifier()
			var data interface{} = tt.data
			var dataPtr *interface{}

			if tt.name == "nil data pointer" {
				dataPtr = nil
			} else {
				dataPtr = &data
			}

			err := m.Set(dataPtr, tt.path, tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(data, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, data)
			}
		})
	}
}

func TestModifierDelete(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		path        string
		expected    interface{}
		expectError bool
	}{
		{
			name:     "delete simple property",
			data:     map[string]interface{}{"name": "John", "age": 30},
			path:     "name",
			expected: map[string]interface{}{"age": 30},
		},
		{
			name:     "delete nested property",
			data:     map[string]interface{}{"user": map[string]interface{}{"name": "John", "age": 30}},
			path:     "user.name",
			expected: map[string]interface{}{"user": map[string]interface{}{"age": 30}},
		},
		{
			name:     "delete non-existent property",
			data:     map[string]interface{}{"name": "John"},
			path:     "age",
			expected: map[string]interface{}{"name": "John"},
		},
		{
			name:        "delete from nil data",
			data:        nil,
			path:        "field",
			expectError: true,
		},
		{
			name:        "delete root",
			data:        map[string]interface{}{"name": "John"},
			path:        "",
			expectError: true,
		},
		{
			name:        "delete with invalid path",
			data:        map[string]interface{}{"name": "John"},
			path:        "",
			expectError: true,
		},
		{
			name:        "delete with nil data pointer",
			data:        map[string]interface{}{"name": "John"},
			path:        "name",
			expectError: true,
		},
		{
			name:        "array element deletion not implemented",
			data:        map[string]interface{}{"items": []interface{}{"a", "b", "c"}},
			path:        "items.1",
			expectError: true,
		},
		{
			name:        "invalid array index for deletion",
			data:        map[string]interface{}{"items": []interface{}{"a", "b"}},
			path:        "items.abc",
			expectError: true,
		},
		{
			name:        "navigate through non-container for deletion",
			data:        map[string]interface{}{"value": "string"},
			path:        "value.subfield",
			expectError: true,
		},
		{
			name:        "delete from non-container",
			data:        "string",
			path:        "field",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModifier()
			var data interface{} = tt.data
			var dataPtr *interface{}

			if tt.name == "delete with nil data pointer" {
				dataPtr = nil
			} else {
				dataPtr = &data
			}

			err := m.Delete(dataPtr, tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(data, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, data)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: nil,
		},
		{
			name:     "simple path",
			path:     "name",
			expected: []string{"name"},
		},
		{
			name:     "nested path",
			path:     "user.name",
			expected: []string{"user", "name"},
		},
		{
			name:     "deep nested path",
			path:     "user.profile.name",
			expected: []string{"user", "profile", "name"},
		},
		{
			name:     "path with array index",
			path:     "items.0.name",
			expected: []string{"items", "0", "name"},
		},
		{
			name:     "path with empty segments",
			path:     "user..name",
			expected: []string{"user", "name"},
		},
		{
			name:     "path starting with dot",
			path:     ".name",
			expected: []string{"name"},
		},
		{
			name:     "path ending with dot",
			path:     "name.",
			expected: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModifier()
			result := m.parsePath(tt.path)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSetAtPathEdgeCases(t *testing.T) {
	m := NewModifier()

	// Test setting with empty segments
	var data interface{} = map[string]interface{}{}
	err := m.setAtPath(&data, []string{}, "value")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if data != "value" {
		t.Errorf("Expected 'value', got %v", data)
	}

	// Test array index bounds
	data = map[string]interface{}{"items": []interface{}{"a"}}
	err = m.setAtPath(&data, []string{"items", "-1"}, "value")
	if err == nil {
		t.Error("Expected error for negative array index")
	}
}

func TestDeleteAtPathEdgeCases(t *testing.T) {
	m := NewModifier()

	// Test deleting with empty segments
	var data interface{} = map[string]interface{}{}
	err := m.deleteAtPath(&data, []string{})
	if err == nil {
		t.Error("Expected error for empty segments")
	}

	// Test array index bounds for deletion
	data = map[string]interface{}{"items": []interface{}{"a"}}
	err = m.deleteAtPath(&data, []string{"items", "-1"})
	if err == nil {
		t.Error("Expected error for negative array index")
	}
}
