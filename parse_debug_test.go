package xjson

import (
	"testing"

	"github.com/474420502/xjson/internal/parser"
)

func TestBooleanLiteralParsing(t *testing.T) {
	// Test parsing of boolean literal
	query := "products[?(@.inStock == true)]"

	p := parser.NewParser(query)
	parsedQuery, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	if len(parsedQuery.Steps) == 0 {
		t.Fatal("No steps parsed")
	}

	step := parsedQuery.Steps[0]
	if len(step.Predicates) == 0 {
		t.Fatal("No predicates parsed")
	}

	predicate := step.Predicates[0]
	if predicate.Type != parser.PredicateExpression {
		t.Fatalf("Expected predicate expression, got %d", predicate.Type)
	}

	expr := predicate.Expression
	if expr.Type != parser.ExpressionBinary {
		t.Fatalf("Expected binary expression, got %d", expr.Type)
	}

	if expr.Operator != "==" {
		t.Errorf("Expected operator ==, got %s", expr.Operator)
	}
	if expr.Left.Type != parser.ExpressionPath {
		t.Errorf("Expected left type ExpressionPath, got %d", expr.Left.Type)
	}
	if expr.Right.Type != parser.ExpressionLiteral {
		t.Errorf("Expected right type ExpressionLiteral, got %d", expr.Right.Type)
	}
	if expr.Right.Value != true {
		t.Errorf("Expected right value true, got %v", expr.Right.Value)
	}
}
