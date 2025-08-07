package path

import (
	"reflect"
	"testing"
)

func TestGetValueBySimplePath(t *testing.T) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": map[string]interface{}{
				"title":  "Test Book",
				"author": "Test Author",
			},
		},
		"numbers": []interface{}{1, 2, 3},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
		found    bool
	}{
		{
			name:     "nested object access",
			path:     "store.book.title",
			expected: "Test Book",
			found:    true,
		},
		{
			name:  "non-existent path",
			path:  "missing.key",
			found: false,
		},
		{
			name:     "root level access",
			path:     "store",
			expected: data["store"],
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := GetValueBySimplePath(data, tt.path)

			if found != tt.found {
				t.Errorf("Expected found=%t, got found=%t", tt.found, found)
				return
			}

			if tt.found && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
