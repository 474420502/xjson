package engine

import (
	"encoding/json"
	"testing"

	"github.com/474420502/xjson/internal/parser"
)

func TestEngine(t *testing.T) {
	jsonData := `{
		"store": {
			"book": [
				{ "category": "reference", "author": "Nigel Rees", "price": 8.95 },
				{ "category": "fiction", "author": "Evelyn Waugh", "price": 12.99 }
			],
			"bicycle": { "color": "red", "price": 19.95 }
		}
	}`
	var data interface{}
	json.Unmarshal([]byte(jsonData), &data)

	engine := NewEngine()

	// Test simple path
	query, _ := parser.NewParser("/store/bicycle/color").Parse()
	matches, err := engine.ExecuteOnMaterialized(data, query)
	if err != nil {
		t.Fatalf("Failed to execute simple path query: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(matches))
	}
	if matches[0].Value != "red" {
		t.Errorf("Expected 'red', got %v", matches[0].Value)
	}

	// Test path with array index
	query, _ = parser.NewParser("/store/book[1]/author").Parse()
	matches, err = engine.ExecuteOnMaterialized(data, query)
	if err != nil {
		t.Fatalf("Failed to execute path with index: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(matches))
	}
	if matches[0].Value != "Nigel Rees" {
		t.Errorf("Expected 'Nigel Rees', got %v", matches[0].Value)
	}
}
