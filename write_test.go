package xjson

import (
	"testing"
)

func TestDocumentSet(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		value    interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "set simple property",
			json:     `{"name": "John", "age": 30}`,
			path:     "name",
			value:    "Jane",
			expected: `{"age":30,"name":"Jane"}`,
			wantErr:  false,
		},
		{
			name:     "set nested property",
			json:     `{"user": {"name": "John", "age": 30}}`,
			path:     "user.name",
			value:    "Jane",
			expected: `{"user":{"age":30,"name":"Jane"}}`,
			wantErr:  false,
		},
		{
			name:     "set new property",
			json:     `{"name": "John"}`,
			path:     "age",
			value:    25,
			expected: `{"age":25,"name":"John"}`,
			wantErr:  false,
		},
		{
			name:     "set new nested property",
			json:     `{"user": {"name": "John"}}`,
			path:     "user.age",
			value:    30,
			expected: `{"user":{"age":30,"name":"John"}}`,
			wantErr:  false,
		},
		{
			name:     "set property in empty object",
			json:     `{}`,
			path:     "name",
			value:    "John",
			expected: `{"name":"John"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.json)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			err = doc.Set(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check if document was materialized
				if !doc.IsMaterialized() {
					t.Error("Document should be materialized after Set operation")
				}

				// Get the modified JSON
				result, err := doc.Bytes()
				if err != nil {
					t.Fatalf("Document.Bytes() error = %v", err)
				}

				// Parse both expected and actual to compare structure
				expectedDoc, err := ParseString(tt.expected)
				if err != nil {
					t.Fatalf("ParseString(expected) error = %v", err)
				}

				actualDoc, err := ParseString(string(result))
				if err != nil {
					t.Fatalf("ParseString(result) error = %v", err)
				}

				// Compare by converting both to raw interface{} and comparing
				if !compareJSON(expectedDoc.materialized, actualDoc.materialized) {
					t.Errorf("Document.Set() result mismatch\nExpected: %s\nActual: %s", tt.expected, string(result))
				}
			}
		})
	}
}

func TestDocumentDelete(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "delete simple property",
			json:     `{"name": "John", "age": 30}`,
			path:     "name",
			expected: `{"age":30}`,
			wantErr:  false,
		},
		{
			name:     "delete nested property",
			json:     `{"user": {"name": "John", "age": 30}}`,
			path:     "user.name",
			expected: `{"user":{"age":30}}`,
			wantErr:  false,
		},
		{
			name:     "delete non-existent property",
			json:     `{"name": "John"}`,
			path:     "age",
			expected: `{"name":"John"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.json)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			err = doc.Delete(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check if document was materialized
				if !doc.IsMaterialized() {
					t.Error("Document should be materialized after Delete operation")
				}

				// Get the modified JSON
				result, err := doc.Bytes()
				if err != nil {
					t.Fatalf("Document.Bytes() error = %v", err)
				}

				// Parse both expected and actual to compare structure
				expectedDoc, err := ParseString(tt.expected)
				if err != nil {
					t.Fatalf("ParseString(expected) error = %v", err)
				}

				actualDoc, err := ParseString(string(result))
				if err != nil {
					t.Fatalf("ParseString(result) error = %v", err)
				}

				// Compare by converting both to raw interface{} and comparing
				if !compareJSON(expectedDoc.materialized, actualDoc.materialized) {
					t.Errorf("Document.Delete() result mismatch\nExpected: %s\nActual: %s", tt.expected, string(result))
				}
			}
		})
	}
}

func TestMaterializeOnWrite(t *testing.T) {
	doc, err := ParseString(`{"name": "John", "age": 30}`)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	// Initially should not be materialized
	if doc.IsMaterialized() {
		t.Error("Document should not be materialized initially")
	}

	// Read operations should not trigger materialization
	result := doc.Query("name")
	if result.Exists() {
		name := result.MustString()
		if name != "John" {
			t.Errorf("Expected name to be 'John', got '%s'", name)
		}
	}

	// Should still not be materialized after read
	if doc.IsMaterialized() {
		t.Error("Document should not be materialized after read operations")
	}

	// Write operation should trigger materialization
	err = doc.Set("name", "Jane")
	if err != nil {
		t.Fatalf("Document.Set() error = %v", err)
	}

	// Should now be materialized
	if !doc.IsMaterialized() {
		t.Error("Document should be materialized after write operation")
	}

	// Verify the change
	result = doc.Query("name")
	if result.Exists() {
		name := result.MustString()
		if name != "Jane" {
			t.Errorf("Expected name to be 'Jane', got '%s'", name)
		}
	}
}

// Helper function to compare JSON structures
func compareJSON(a, b interface{}) bool {
	// Simple deep comparison - in a real implementation you might want more sophisticated comparison
	switch va := a.(type) {
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if bv, exists := vb[k]; !exists || !compareJSON(v, bv) {
				return false
			}
		}
		return true
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(va) != len(vb) {
			return false
		}
		for i, v := range va {
			if !compareJSON(v, vb[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
