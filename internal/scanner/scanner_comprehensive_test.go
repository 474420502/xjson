package scanner

import (
	"testing"
)

func TestNewScanner(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	scanner := NewScanner(data)

	if scanner == nil {
		t.Error("NewScanner should not return nil")
	}
	if scanner.Position() != 0 {
		t.Errorf("Initial position should be 0, got %d", scanner.Position())
	}
}

func TestScannerBasicOperations(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	scanner := NewScanner(data)

	// Test Current()
	current := scanner.Current()
	if current != '{' {
		t.Errorf("Expected '{', got '%c'", current)
	}

	// Test Advance()
	scanner.Advance(1)
	current = scanner.Current()
	if current != '"' {
		t.Errorf("Expected '\"', got '%c'", current)
	}

	// Test Peek()
	peeked := scanner.Peek(1)
	if peeked != 'k' {
		t.Errorf("Expected 'k', got '%c'", peeked)
	}

	// Peek should not advance position
	current = scanner.Current()
	if current != '"' {
		t.Errorf("Peek should not advance position, still expect '\"', got '%c'", current)
	}

	// Test Reset()
	scanner.Reset()
	current = scanner.Current()
	if current != '{' {
		t.Errorf("After reset, expected '{', got '%c'", current)
	}
}

func TestScannerBoundaryConditions(t *testing.T) {
	data := []byte(`abc`)
	scanner := NewScanner(data)

	// Advance to end
	scanner.Advance(3)
	current := scanner.Current()
	if current != 0 {
		t.Errorf("At end of data, expected 0, got %d", current)
	}

	// Advance beyond end
	scanner.Advance(10)
	current = scanner.Current()
	if current != 0 {
		t.Errorf("Beyond end of data, expected 0, got %d", current)
	}

	// Peek beyond end
	scanner.Reset()
	peeked := scanner.Peek(10)
	if peeked != 0 {
		t.Errorf("Peek beyond end should return 0, got %d", peeked)
	}
}

func TestScannerEmptyData(t *testing.T) {
	scanner := NewScanner([]byte{})

	current := scanner.Current()
	if current != 0 {
		t.Errorf("Empty data should return 0, got %d", current)
	}

	peeked := scanner.Peek(0)
	if peeked != 0 {
		t.Errorf("Peek on empty data should return 0, got %d", peeked)
	}

	scanner.Advance(1)
	current = scanner.Current()
	if current != 0 {
		t.Errorf("Advance on empty data should still return 0, got %d", current)
	}
}

func TestScannerSkipWhitespace(t *testing.T) {
	data := []byte(`   	  {"key": "value"}`)
	scanner := NewScanner(data)

	scanner.SkipWhitespace()
	current := scanner.Current()
	if current != '{' {
		t.Errorf("After skipping whitespace, expected '{', got '%c'", current)
	}

	// Test with no whitespace
	data2 := []byte(`{"key": "value"}`)
	scanner2 := NewScanner(data2)
	scanner2.SkipWhitespace()
	current = scanner2.Current()
	if current != '{' {
		t.Errorf("With no whitespace, expected '{', got '%c'", current)
	}

	// Test with only whitespace
	data3 := []byte(`   	  `)
	scanner3 := NewScanner(data3)
	scanner3.SkipWhitespace()
	current = scanner3.Current()
	if current != 0 {
		t.Errorf("With only whitespace, expected 0, got %d", current)
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		expected   string
		shouldWork bool
	}{
		{
			name:       "simple string",
			data:       `"hello"`,
			expected:   "hello",
			shouldWork: true,
		},
		{
			name:       "string with escapes",
			data:       `"hello\nworld"`,
			expected:   "hello\\nworld", // ReadString might return raw string
			shouldWork: true,
		},
		{
			name:       "empty string",
			data:       `""`,
			expected:   "",
			shouldWork: true,
		},
		{
			name:       "unterminated string",
			data:       `"hello`,
			shouldWork: false,
		},
		{
			name:       "no quotes",
			data:       `hello`,
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.data))
			result, ok := scanner.ReadString()

			if tt.shouldWork {
				if !ok {
					t.Error("Expected ReadString to succeed")
					return
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			} else {
				if ok {
					t.Error("Expected ReadString to fail")
				}
			}
		})
	}
}

func TestReadNumber(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		expected   string
		shouldWork bool
	}{
		{
			name:       "integer",
			data:       "123",
			expected:   "123",
			shouldWork: true,
		},
		{
			name:       "negative integer",
			data:       "-123",
			expected:   "-123",
			shouldWork: true,
		},
		{
			name:       "float",
			data:       "123.456",
			expected:   "123.456",
			shouldWork: true,
		},
		{
			name:       "scientific notation",
			data:       "1.23e10",
			expected:   "1.23e10",
			shouldWork: true,
		},
		{
			name:       "zero",
			data:       "0",
			expected:   "0",
			shouldWork: true,
		},
		{
			name:       "invalid - not a number",
			data:       "abc",
			shouldWork: false,
		},
		{
			name:       "invalid - just minus",
			data:       "-",
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.data))
			result, ok := scanner.ReadNumber()

			if tt.shouldWork {
				if !ok {
					t.Error("Expected ReadNumber to succeed")
					return
				}
				if string(result) != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, string(result))
				}
			} else {
				if ok {
					t.Error("Expected ReadNumber to fail")
				}
			}
		})
	}
}

func TestReadBool(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		expected   bool
		shouldWork bool
	}{
		{
			name:       "true",
			data:       "true",
			expected:   true,
			shouldWork: true,
		},
		{
			name:       "false",
			data:       "false",
			expected:   false,
			shouldWork: true,
		},
		{
			name:       "invalid",
			data:       "maybe",
			shouldWork: false,
		},
		{
			name:       "partial true",
			data:       "tru",
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.data))
			result, ok := scanner.ReadBool()

			if tt.shouldWork {
				if !ok {
					t.Error("Expected ReadBool to succeed")
					return
				}
				if result != tt.expected {
					t.Errorf("Expected %t, got %t", tt.expected, result)
				}
			} else {
				if ok {
					t.Error("Expected ReadBool to fail")
				}
			}
		})
	}
}

func TestReadNull(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		shouldWork bool
	}{
		{
			name:       "null",
			data:       "null",
			shouldWork: true,
		},
		{
			name:       "not null",
			data:       "nope",
			shouldWork: false,
		},
		{
			name:       "partial null",
			data:       "nul",
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.data))
			ok := scanner.ReadNull()

			if tt.shouldWork && !ok {
				t.Error("Expected ReadNull to succeed")
			} else if !tt.shouldWork && ok {
				t.Error("Expected ReadNull to fail")
			}
		})
	}
}

func TestSkipValue(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "string value",
			data: `"hello world"`,
		},
		{
			name: "number value",
			data: `123.456`,
		},
		{
			name: "boolean value",
			data: `true`,
		},
		{
			name: "null value",
			data: `null`,
		},
		{
			name: "simple object",
			data: `{"key": "value"}`,
		},
		{
			name: "simple array",
			data: `[1, 2, 3]`,
		},
		{
			name: "nested object",
			data: `{"a": {"b": {"c": 123}}}`,
		},
		{
			name: "nested array",
			data: `[1, [2, [3, 4]], 5]`,
		},
		{
			name: "complex nested structure",
			data: `{"users": [{"name": "John", "items": [1, 2, 3]}, {"name": "Jane"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.data))

			ok := scanner.SkipValue()
			if !ok {
				t.Errorf("SkipValue should succeed for: %s", tt.data)
			}

			// Should have consumed the entire value
			if !scanner.IsEOF() {
				t.Errorf("Expected to consume entire value, but position is %d", scanner.Position())
			}
		})
	}
}

func TestScannerWithInvalidJSON(t *testing.T) {
	invalidCases := []string{
		`{`,             // Unclosed object
		`[`,             // Unclosed array
		`{"key": "val"`, // Unclosed string
	}

	for _, data := range invalidCases {
		t.Run(data, func(t *testing.T) {
			scanner := NewScanner([]byte(data))
			ok := scanner.SkipValue()
			if ok {
				t.Errorf("Expected SkipValue to fail for invalid JSON: %s", data)
			}
		})
	}
}

func TestPosition(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	scanner := NewScanner(data)

	if scanner.Position() != 0 {
		t.Errorf("Initial position should be 0, got %d", scanner.Position())
	}

	scanner.Advance(5)
	if scanner.Position() != 5 {
		t.Errorf("Position after advancing 5 should be 5, got %d", scanner.Position())
	}

	scanner.Reset()
	if scanner.Position() != 0 {
		t.Errorf("Position after reset should be 0, got %d", scanner.Position())
	}
}

func TestIsEOF(t *testing.T) {
	data := []byte(`abc`)
	scanner := NewScanner(data)

	if scanner.IsEOF() {
		t.Error("Should not be EOF at start")
	}

	scanner.Advance(3)
	if !scanner.IsEOF() {
		t.Error("Should be EOF after advancing to end")
	}

	scanner.Advance(10) // Advance beyond end
	if !scanner.IsEOF() {
		t.Error("Should still be EOF after advancing beyond end")
	}
}

func TestSetPosition(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	scanner := NewScanner(data)

	scanner.SetPosition(6)
	if scanner.Position() != 6 {
		t.Errorf("SetPosition(6) should set position to 6, got %d", scanner.Position())
	}

	current := scanner.Current()
	if current != ':' {
		t.Errorf("At position 6, expected ':', got '%c'", current)
	}
}

func TestRemaining(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	scanner := NewScanner(data)

	initialRemaining := scanner.Remaining()
	if initialRemaining != len(data) {
		t.Errorf("Initial remaining should be %d, got %d", len(data), initialRemaining)
	}

	scanner.Advance(5)
	remaining := scanner.Remaining()
	expected := len(data) - 5
	if remaining != expected {
		t.Errorf("After advancing 5, remaining should be %d, got %d", expected, remaining)
	}
}

func TestFindKey(t *testing.T) {
	data := []byte(`{"name": "John", "age": 30, "city": "NYC"}`)
	scanner := NewScanner(data)

	// Should find existing key
	found := scanner.FindKey("age")
	if !found {
		t.Error("Should find existing key 'age'")
	}

	// Reset and try non-existing key
	scanner.Reset()
	found = scanner.FindKey("country")
	if found {
		t.Error("Should not find non-existing key 'country'")
	}

	// Test with nested object
	data2 := []byte(`{"user": {"name": "John", "profile": {"age": 30}}}`)
	scanner2 := NewScanner(data2)
	found = scanner2.FindKey("user")
	if !found {
		t.Error("Should find 'user' key in nested object")
	}
}

func TestGetValueAt(t *testing.T) {
	data := []byte(`{"name": "John", "age": 30}`)
	scanner := NewScanner(data)

	// Find a key and get its value
	if scanner.FindKey("name") {
		value, ok := scanner.GetValueAt()
		if !ok {
			t.Error("Should be able to get value after finding key")
		}
		if string(value) != `"John"` {
			t.Errorf("Expected '\"John\"', got '%s'", string(value))
		}
	}

	// Reset and try with number value
	scanner.Reset()
	if scanner.FindKey("age") {
		value, ok := scanner.GetValueAt()
		if !ok {
			t.Error("Should be able to get number value")
		}
		if string(value) != "30" {
			t.Errorf("Expected '30', got '%s'", string(value))
		}
	}
}
