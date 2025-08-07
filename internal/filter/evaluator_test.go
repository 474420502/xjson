package filter

import (
	"testing"

	"github.com/474420502/xjson/internal/parser"
)

func TestEvaluateLiteralExpression(t *testing.T) {
	fe := NewFilterEvaluator()
	cases := []struct {
		val    interface{}
		expect bool
	}{
		{true, true},
		{false, false},
		{1.0, true},
		{0.0, false},
		{"", false},
		{"abc", true},
	}
	for _, c := range cases {
		expr := &parser.Expression{Type: parser.ExpressionLiteral, Value: c.val}
		got, _ := fe.evaluateLiteralExpression(expr, nil)
		if got != c.expect {
			t.Errorf("Literal %v: expect %v, got %v", c.val, c.expect, got)
		}
	}
}

func TestEvaluateBinaryExpression(t *testing.T) {
	fe := NewFilterEvaluator()
	ctx := &EvaluationContext{CurrentItem: map[string]interface{}{"x": 2}}
	left := &parser.Expression{Type: parser.ExpressionPath, Path: []string{"x"}}
	right := &parser.Expression{Type: parser.ExpressionLiteral, Value: 2}
	expr := &parser.Expression{Type: parser.ExpressionBinary, Operator: "==", Left: left, Right: right}
	ok, err := fe.evaluateBinaryExpression(expr, ctx)
	if err != nil || !ok {
		t.Errorf("Binary == failed: %v, %v", ok, err)
	}
}

func TestEvaluateFunctionExpression_exists(t *testing.T) {
	fe := NewFilterEvaluator()
	ctx := &EvaluationContext{CurrentItem: map[string]interface{}{"foo": 1}}
	expr := &parser.Expression{
		Type:     parser.ExpressionFunction,
		Function: "exists",
		Left: &parser.Expression{
			Type: parser.ExpressionPath,
			Path: []string{"foo"},
		},
	}
	ok, err := fe.evaluateFunctionExpression(expr, ctx)
	if err != nil || !ok {
		t.Errorf("exists() should be true, got %v, %v", ok, err)
	}
}

// Test EvaluateExpression function (0% coverage)
func TestEvaluateExpression(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		expr     *parser.Expression
		ctx      *EvaluationContext
		expected bool
		hasError bool
	}{
		{
			name: "literal true",
			expr: &parser.Expression{
				Type:  parser.ExpressionLiteral,
				Value: true,
			},
			ctx:      &EvaluationContext{},
			expected: true,
		},
		{
			name: "literal false",
			expr: &parser.Expression{
				Type:  parser.ExpressionLiteral,
				Value: false,
			},
			ctx:      &EvaluationContext{},
			expected: false,
		},
		{
			name: "binary expression",
			expr: &parser.Expression{
				Type:     parser.ExpressionBinary,
				Operator: "==",
				Left: &parser.Expression{
					Type:  parser.ExpressionLiteral,
					Value: 1,
				},
				Right: &parser.Expression{
					Type:  parser.ExpressionLiteral,
					Value: 1,
				},
			},
			ctx:      &EvaluationContext{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fe.EvaluateExpression(tt.expr, tt.ctx)

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

			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

// Test evaluatePathExpression function (0% coverage)
func TestEvaluatePathExpression(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		expr     *parser.Expression
		ctx      *EvaluationContext
		expected bool
		hasError bool
	}{
		{
			name: "path exists",
			expr: &parser.Expression{
				Type: parser.ExpressionPath,
				Path: []string{"name"},
			},
			ctx: &EvaluationContext{
				CurrentItem: map[string]interface{}{
					"name": "test",
				},
			},
			expected: true,
		},
		{
			name: "path doesn't exist",
			expr: &parser.Expression{
				Type: parser.ExpressionPath,
				Path: []string{"missing"},
			},
			ctx: &EvaluationContext{
				CurrentItem: map[string]interface{}{
					"name": "test",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fe.evaluatePathExpression(tt.expr, tt.ctx)

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

			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

// Test evaluateUnaryExpression function (0% coverage)
func TestEvaluateUnaryExpression(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		expr     *parser.Expression
		ctx      *EvaluationContext
		expected bool
		hasError bool
	}{
		{
			name: "NOT true",
			expr: &parser.Expression{
				Type:     parser.ExpressionUnary,
				Operator: "!",
				Left: &parser.Expression{
					Type:  parser.ExpressionLiteral,
					Value: true,
				},
			},
			ctx:      &EvaluationContext{},
			expected: false,
		},
		{
			name: "NOT false",
			expr: &parser.Expression{
				Type:     parser.ExpressionUnary,
				Operator: "!",
				Left: &parser.Expression{
					Type:  parser.ExpressionLiteral,
					Value: false,
				},
			},
			ctx:      &EvaluationContext{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fe.evaluateUnaryExpression(tt.expr, tt.ctx)

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

			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

// Test more comparison operators
func TestComparisons(t *testing.T) {
	fe := NewFilterEvaluator()

	tests := []struct {
		name     string
		left     interface{}
		operator string
		right    interface{}
		expected bool
	}{
		{"equals string", "test", "==", "test", true},
		{"not equals string", "test", "!=", "other", true},
		{"less than", 1, "<", 2, true},
		{"less than or equal", 1, "<=", 1, true},
		{"greater than", 2, ">", 1, true},
		{"greater than or equal", 2, ">=", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := &parser.Expression{Type: parser.ExpressionLiteral, Value: tt.left}
			right := &parser.Expression{Type: parser.ExpressionLiteral, Value: tt.right}
			expr := &parser.Expression{
				Type:     parser.ExpressionBinary,
				Operator: tt.operator,
				Left:     left,
				Right:    right,
			}

			result, err := fe.evaluateBinaryExpression(expr, &EvaluationContext{})
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestValuesEqual(t *testing.T) {
	fe := NewFilterEvaluator()
	cases := []struct {
		a, b interface{}
		eq   bool
	}{
		{1, 1.0, true},
		{"a", "a", true},
		{true, 1, true},
		{false, 0, true},
		{nil, nil, true},
		{1, 2, false},
	}
	for _, c := range cases {
		if fe.valuesEqual(c.a, c.b) != c.eq {
			t.Errorf("valuesEqual(%v, %v) expect %v", c.a, c.b, c.eq)
		}
	}
}

func TestCompareValues(t *testing.T) {
	fe := NewFilterEvaluator()
	cases := []struct {
		l, r interface{}
		op   string
		exp  bool
	}{
		{1, 1, "==", true},
		{1, 2, "!=", true},
		{2, 1, ">", true},
		{1, 2, "<", true},
		{2, 2, ">=", true},
		{1, 1, "<=", true},
	}
	for _, c := range cases {
		got, _ := fe.compareValues(c.l, c.r, c.op)
		if got != c.exp {
			t.Errorf("compareValues(%v,%v,%s) expect %v, got %v", c.l, c.r, c.op, c.exp, got)
		}
	}
}

func TestToBool(t *testing.T) {
	fe := NewFilterEvaluator()
	cases := []struct {
		v   interface{}
		exp bool
	}{
		{true, true},
		{false, false},
		{1, true},
		{0, false},
		{"", false},
		{"x", true},
		{nil, false},
	}
	for _, c := range cases {
		if fe.toBool(c.v) != c.exp {
			t.Errorf("toBool(%v) expect %v", c.v, c.exp)
		}
	}
}
