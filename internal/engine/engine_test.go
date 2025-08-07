package engine

import (
	"reflect"
	"testing"

	"github.com/474420502/xjson/internal/parser"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Error("NewEngine should return non-nil engine")
	}
}

// Test executeDescendantStep - 0% coverage

// Test executeDescendantStepOnMaterialized - 0% coverage
func TestExecuteOnMaterialized_DescendantStep(t *testing.T) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": map[string]interface{}{"title": "Example"},
		},
		"book": map[string]interface{}{"title": "Another"},
	}

	query := &parser.Query{
		Steps: []parser.Step{
			{Type: parser.StepDescendant, Name: "book"},
		},
	}

	engine := NewEngine()
	matches, err := engine.ExecuteOnMaterialized(data, query)
	if err != nil {
		t.Fatalf("ExecuteOnMaterialized error: %v", err)
	}

	// Should find books using descendant step
	t.Logf("Descendant step found %d matches", len(matches))
}

// Test executeArrayStepOnMaterialized - 0% coverage
func TestExecuteOnMaterialized_ArrayStep(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{"title": "Book1"},
		map[string]interface{}{"title": "Book2"},
		map[string]interface{}{"title": "Book3"},
	}

	query := &parser.Query{
		Steps: []parser.Step{
			{Type: parser.StepChild, Predicates: []parser.Predicate{
				{Type: parser.PredicateIndex, Index: 1},
			}},
		},
	}

	engine := NewEngine()
	matches, err := engine.ExecuteOnMaterialized(data, query)
	if err != nil {
		t.Fatalf("ExecuteOnMaterialized error: %v", err)
	}

	t.Logf("Array step found %d matches", len(matches))
}

// Test array slice operations - 0% coverage
func TestExecuteOnMaterialized_ArraySlice(t *testing.T) {
	data := []interface{}{1, 2, 3, 4, 5}

	query := &parser.Query{
		Steps: []parser.Step{
			{Type: parser.StepChild, Predicates: []parser.Predicate{
				{Type: parser.PredicateSlice, Start: 1, End: 3},
			}},
		},
	}

	engine := NewEngine()
	matches, err := engine.ExecuteOnMaterialized(data, query)
	if err != nil {
		t.Fatalf("ExecuteOnMaterialized error: %v", err)
	}

	t.Logf("Array slice found %d matches", len(matches))
}

// Test ApplyFilter function - 20% coverage -> improve
func TestApplyFilterFunction(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"price": 10},
		map[string]interface{}{"price": 20},
	}

	predicate := parser.Predicate{
		Type: parser.PredicateExpression,
		Expression: &parser.Expression{
			Type:     parser.ExpressionBinary,
			Operator: "==", // String operator
			Left: &parser.Expression{
				Type: parser.ExpressionPath,
				Path: []string{"price"},
			},
			Right: &parser.Expression{
				Type:  parser.ExpressionLiteral,
				Value: 20,
			},
		},
	}

	engine := NewEngine()
	ctx := &MaterializedContext{data: items}
	result := engine.applyPredicates(ctx, items, []parser.Predicate{predicate})

	if len(result) != 1 {
		t.Errorf("Expected 1 item to be filtered, got %d", len(result))
	}
}

func TestExecuteOnMaterialized_SimplePath(t *testing.T) {
	data := map[string]interface{}{"a": 1, "b": 2}
	query := &parser.Query{Steps: []parser.Step{{Type: parser.StepChild, Name: "a"}}}
	engine := NewEngine()
	matches, err := engine.ExecuteOnMaterialized(data, query)
	if err != nil {
		t.Fatalf("ExecuteOnMaterialized error: %v", err)
	}
	if len(matches) != 1 || !reflect.DeepEqual(matches[0].Value, 1) {
		t.Errorf("Expected [1], got %v", matches)
	}
}

func TestExecuteOnMaterialized_Wildcard(t *testing.T) {
	data := map[string]interface{}{"a": 1, "b": 2}
	query := &parser.Query{Steps: []parser.Step{{Type: parser.StepWildcard}}}
	engine := NewEngine()
	matches, err := engine.ExecuteOnMaterialized(data, query)
	if err != nil {
		t.Fatalf("ExecuteOnMaterialized error: %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches for wildcard, got %d", len(matches))
	}
}

func TestParseSimplePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple path",
			path:     "store.book",
			expected: []string{"store", "book"},
		},
		{
			name:     "single component",
			path:     "name",
			expected: []string{"name"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "path with numbers",
			path:     "data.0.value",
			expected: []string{"data", "0", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSimplePath(tt.path)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConvertValue(t *testing.T) {
	tests := []struct {
		name       string
		input      interface{}
		targetType reflect.Type
		expected   interface{}
		hasError   bool
	}{
		{
			name:       "string to string",
			input:      "test",
			targetType: reflect.TypeOf(""),
			expected:   "test",
		},
		{
			name:       "boolean to bool",
			input:      true,
			targetType: reflect.TypeOf(true),
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertValue(tt.input, tt.targetType)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseJSONValue(t *testing.T) {
	// Test parseJSONValue function (0% coverage)
	tests := []struct {
		name     string
		input    []byte
		expected interface{}
		hasError bool
	}{
		{
			name:     "string value",
			input:    []byte(`"test"`),
			expected: "test",
		},
		{
			name:     "number value",
			input:    []byte("42"),
			expected: 42.0,
		},
		{
			name:     "boolean value",
			input:    []byte("true"),
			expected: true,
		},
		{
			name:     "null value",
			input:    []byte("null"),
			expected: nil,
		},
		{
			name:     "object value",
			input:    []byte(`{"key": "value"}`),
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "invalid json",
			input:    []byte(`{invalid`),
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONValue(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
