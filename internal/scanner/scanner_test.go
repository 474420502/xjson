package scanner

import (
	"testing"
)

func TestScannerBasics(t *testing.T) {
	data := []byte(`{"name": "test", "value": 42}`)
	scanner := NewScanner(data)

	if scanner.Current() != '{' {
		t.Errorf("Expected '{', got %c", scanner.Current())
	}

	if scanner.Position() != 0 {
		t.Errorf("Expected position 0, got %d", scanner.Position())
	}

	if scanner.Remaining() != len(data) {
		t.Errorf("Expected remaining %d, got %d", len(data), scanner.Remaining())
	}
}

func TestSkipWhitespace(t *testing.T) {
	data := []byte(`   {"name": "test"}`)
	scanner := NewScanner(data)

	scanner.SkipWhitespace()

	if scanner.Current() != '{' {
		t.Errorf("Expected '{', got %c", scanner.Current())
	}

	if scanner.Position() != 3 {
		t.Errorf("Expected position 3, got %d", scanner.Position())
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		success  bool
	}{
		{
			name:     "simple string",
			input:    `"hello"`,
			expected: "hello",
			success:  true,
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
			success:  true,
		},
		{
			name:     "string with spaces",
			input:    `"hello world"`,
			expected: "hello world",
			success:  true,
		},
		{
			name:     "unterminated string",
			input:    `"hello`,
			expected: "",
			success:  false,
		},
		{
			name:     "not a string",
			input:    `123`,
			expected: "",
			success:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.input))
			result, ok := scanner.ReadString()

			if ok != tt.success {
				t.Errorf("Expected success %v, got %v", tt.success, ok)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestReadNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		success  bool
	}{
		{
			name:     "integer",
			input:    `123`,
			expected: "123",
			success:  true,
		},
		{
			name:     "negative integer",
			input:    `-123`,
			expected: "-123",
			success:  true,
		},
		{
			name:     "float",
			input:    `123.45`,
			expected: "123.45",
			success:  true,
		},
		{
			name:     "scientific notation",
			input:    `1.23e10`,
			expected: "1.23e10",
			success:  true,
		},
		{
			name:     "zero",
			input:    `0`,
			expected: "0",
			success:  true,
		},
		{
			name:     "not a number",
			input:    `abc`,
			expected: "",
			success:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.input))
			result, ok := scanner.ReadNumber()

			if ok != tt.success {
				t.Errorf("Expected success %v, got %v", tt.success, ok)
			}

			if string(result) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestReadBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		success  bool
	}{
		{
			name:     "true",
			input:    `true`,
			expected: true,
			success:  true,
		},
		{
			name:     "false",
			input:    `false`,
			expected: false,
			success:  true,
		},
		{
			name:     "not a boolean",
			input:    `123`,
			expected: false,
			success:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.input))
			result, ok := scanner.ReadBool()

			if ok != tt.success {
				t.Errorf("Expected success %v, got %v", tt.success, ok)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFindKey(t *testing.T) {
	data := []byte(`{"name": "test", "value": 42, "nested": {"inner": true}}`)

	tests := []struct {
		name    string
		key     string
		success bool
	}{
		{
			name:    "existing key",
			key:     "name",
			success: true,
		},
		{
			name:    "another existing key",
			key:     "value",
			success: true,
		},
		{
			name:    "non-existing key",
			key:     "missing",
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(data)
			found := scanner.FindKey(tt.key)

			if found != tt.success {
				t.Errorf("Expected %v, got %v", tt.success, found)
			}

			if found && tt.key == "name" {
				// If we found "name", the scanner should be positioned at the string value
				value, ok := scanner.ReadString()
				if !ok || value != "test" {
					t.Errorf("Expected to read 'test', got %q (ok=%v)", value, ok)
				}
			}
		})
	}
}

func BenchmarkScannerReadString(b *testing.B) {
	data := []byte(`"this is a test string with some content"`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner(data)
		_, _ = scanner.ReadString()
	}
}

func BenchmarkScannerReadNumber(b *testing.B) {
	data := []byte(`123.456e10`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner(data)
		_, _ = scanner.ReadNumber()
	}
}

func BenchmarkScannerFindKey(b *testing.B) {
	data := []byte(`{"key1": "value1", "key2": "value2", "key3": "value3", "target": "found"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner(data)
		_ = scanner.FindKey("target")
	}
}
