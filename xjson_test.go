package xjson

import (
	"testing"
)

func TestDocumentParsing(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple object",
			input:   `{"name": "test", "value": 42}`,
			wantErr: false,
		},
		{
			name:    "valid array",
			input:   `[1, 2, 3]`,
			wantErr: false,
		},
		{
			name:    "valid nested object",
			input:   `{"data": {"items": [{"id": 1, "name": "item1"}]}}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"name": "test"`,
			wantErr: true,
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   `[]`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := ParseString(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseString() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseString() unexpected error: %v", err)
				return
			}

			if !doc.IsValid() {
				t.Errorf("ParseString() document should be valid")
			}

			if doc.IsMaterialized() {
				t.Errorf("ParseString() document should not be materialized initially")
			}
		})
	}
}

func TestDocumentBytes(t *testing.T) {
	input := `{"name": "test", "value": 42}`
	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error: %v", err)
	}

	bytes, err := doc.Bytes()
	if err != nil {
		t.Errorf("Bytes() error: %v", err)
	}

	if string(bytes) != input {
		t.Errorf("Bytes() = %s, want %s", string(bytes), input)
	}
}

func TestDocumentString(t *testing.T) {
	input := `{"name": "test", "value": 42}`
	doc, err := ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error: %v", err)
	}

	str, err := doc.String()
	if err != nil {
		t.Errorf("String() error: %v", err)
	}

	if str != input {
		t.Errorf("String() = %s, want %s", str, input)
	}
}

// Benchmark tests to verify lazy parsing performance
func BenchmarkParse(b *testing.B) {
	jsonData := []byte(`{
		"store": {
			"book": [
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "price": 8.99},
				{"category": "fiction", "author": "J.R.R. Tolkien", "title": "The Lord of the Rings", "price": 22.99},
				{"category": "science", "author": "Carl Sagan", "title": "Cosmos", "price": 15.99}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseString(b *testing.B) {
	jsonStr := `{
		"store": {
			"book": [
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "price": 8.99},
				{"category": "fiction", "author": "J.R.R. Tolkien", "title": "The Lord of the Rings", "price": 22.99},
				{"category": "science", "author": "Carl Sagan", "title": "Cosmos", "price": 15.99}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseString(jsonStr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
