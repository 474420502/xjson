package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestNodeSetByPathMethod(t *testing.T) {
	// Test data with nested structure
	jsonData := []byte(`{
		"name": "John Doe",
		"age": 30,
		"isStudent": false,
		"courses": [
			{"title": "Math", "score": 90},
			{"title": "Science", "score": 85},
			{"title": "History", "score": 78}
		],
		"address": {
			"street": "123 Main St",
			"city": "Anytown",
			"zip": "12345"
		},
		"hobbies": ["reading", "hiking", "coding"]
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test SetByPath on existing path
	t.Run("SetByPathExistingPath", func(t *testing.T) {
		result := root.SetByPath("/name", "Jane Smith")
		if result.Error() != nil {
			t.Errorf("SetByPath failed: %v", result.Error())
		}

		name := root.Query("/name")
		if name.String() != "Jane Smith" {
			t.Errorf("Expected 'Jane Smith', got '%s'", name.String())
		}
	})

	// Test SetByPath on nested object
	t.Run("SetByPathNestedObject", func(t *testing.T) {
		result := root.SetByPath("/address/city", "New York")
		if result.Error() != nil {
			t.Errorf("SetByPath failed: %v", result.Error())
		}

		city := root.Query("/address/city")
		if city.String() != "New York" {
			t.Errorf("Expected 'New York', got '%s'", city.String())
		}
	})

	// Test SetByPath on array element
	t.Run("SetByPathArrayElement", func(t *testing.T) {
		result := root.SetByPath("/courses/0/title", "Advanced Math")
		if result.Error() != nil {
			t.Errorf("SetByPath failed: %v", result.Error())
		}

		title := root.Query("/courses/0/title")
		if title.String() != "Advanced Math" {
			t.Errorf("Expected 'Advanced Math', got '%s'", title.String())
		}
	})

	// Test SetByPath add new field to object
	t.Run("SetByPathAddNewField", func(t *testing.T) {
		result := root.SetByPath("/address/country", "USA")
		if result.Error() != nil {
			t.Errorf("SetByPath failed: %v", result.Error())
		}

		country := root.Query("/address/country")
		if country.String() != "USA" {
			t.Errorf("Expected 'USA', got '%s'", country.String())
		}
	})

	// Test SetByPath with invalid path
	t.Run("SetByPathInvalidPath", func(t *testing.T) {
		result := root.SetByPath("/nonexistent/path", "value")
		if result.Error() == nil {
			t.Error("Expected error for invalid path, but got none")
		}
	})

	// Test SetByPath with invalid index
	t.Run("SetByPathInvalidIndex", func(t *testing.T) {
		result := root.SetByPath("/hobbies/10", "value")
		if result.Error() == nil {
			t.Error("Expected error for invalid index, but got none")
		}
	})

	// Test SetByPath with invalid node type
	t.Run("SetByPathInvalidNodeType", func(t *testing.T) {
		result := root.SetByPath("/name/invalid", "value")
		if result.Error() == nil {
			t.Error("Expected error for setting key on string node, but got none")
		}
	})
}

func TestNodeSetOperations(t *testing.T) {
	// Test data with nested structure
	jsonData := []byte(`{
		"name": "John Doe",
		"age": 30,
		"courses": [
			{"title": "Math", "score": 90},
			{"title": "Science", "score": 85}
		],
		"address": {
			"street": "123 Main St",
			"city": "Anytown"
		}
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Run("SetOnObject", func(t *testing.T) {
		address := root.Get("address")
		if address.Type() != core.Object {
			t.Fatalf("Expected object type, got %s", address.Type())
		}

		result := address.Set("country", "USA")
		if result.Error() != nil {
			t.Errorf("Set failed: %v", result.Error())
		}

		country := root.Query("/address/country")
		if country.String() != "USA" {
			t.Errorf("Expected 'USA', got '%s'", country.String())
		}
	})

	t.Run("SetOnArray", func(t *testing.T) {
		courses := root.Get("courses")
		if courses.Type() != core.Array {
			t.Fatalf("Expected array type, got %s", courses.Type())
		}

		newCourse := map[string]interface{}{
			"title": "Physics",
			"score": 92,
		}
		result := courses.Set("0", newCourse)
		if result.Error() != nil {
			t.Errorf("Set failed: %v", result.Error())
		}

		title := root.Query("/courses/0/title")
		if title.String() != "Physics" {
			t.Errorf("Expected 'Physics', got '%s'", title.String())
		}
	})

	t.Run("SetOnInvalidNode", func(t *testing.T) {
		invalidNode := newInvalidNode(nil)
		result := invalidNode.Set("key", "value")
		if result != invalidNode {
			t.Error("Expected same invalid node to be returned")
		}
	})
}