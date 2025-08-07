package parser

import (
	"testing"
)

func TestParser(t *testing.T) {
	// Simple path
	query, err := NewParser("/user/name").Parse()
	if err != nil {
		t.Fatalf("Failed to parse simple path: %v", err)
	}
	if len(query.Steps) != 2 {
		t.Fatalf("Expected 2 steps, got %d", len(query.Steps))
	}
	if query.Steps[0].Name != "user" || query.Steps[1].Name != "name" {
		t.Errorf("Unexpected step names: %v", query.Steps)
	}

	// Path with index
	query, err = NewParser("/users[0]").Parse()
	if err != nil {
		t.Fatalf("Failed to parse path with index: %v", err)
	}
	if len(query.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(query.Steps))
	}
	if len(query.Steps[0].Predicates) != 1 {
		t.Fatalf("Expected 1 predicate, got %d", len(query.Steps[0].Predicates))
	}
	if query.Steps[0].Predicates[0].Index != 0 {
		t.Errorf("Expected index 0, got %d", query.Steps[0].Predicates[0].Index)
	}

	// Path with filter
	query, err = NewParser("//book[@price<10]").Parse()
	if err != nil {
		t.Fatalf("Failed to parse path with filter: %v", err)
	}
	if len(query.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(query.Steps))
	}
	if query.Steps[0].Type != StepDescendant {
		t.Errorf("Expected descendant step, got %v", query.Steps[0].Type)
	}
}
