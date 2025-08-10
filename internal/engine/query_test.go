package engine

import (
	"testing"

	"github.com/474420502/xjson/internal/core"
)

func TestRecursiveDescent(t *testing.T) {
	// Test data with nested structure
	jsonData := []byte(`{
		"store": {
			"books": [
				{"title": "Moby Dick", "price": 8.99, "tags": ["classic", "adventure"]},
				{"title": "Clean Code", "price": 29.99, "tags": ["programming"]}
			],
			"electronics": {
				"phones": [
					{"title": "iPhone", "price": 999.99}
				]
			}
		},
		"prices": [1.99, 5.99, 10.99]
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Test recursive descent to find all "title" fields
	result := root.Query("//title")
	if !result.IsValid() {
		t.Fatalf("Recursive descent query failed: %v", result.Error())
	}

	if result.Type() != core.Array {
		t.Fatalf("Expected array result, got %v", result.Type())
	}

	// Should find 3 titles: "Moby Dick", "Clean Code", "iPhone"
	titles := result.Strings()
	if len(titles) != 3 {
		t.Fatalf("Expected 3 titles, got %d: %v", len(titles), titles)
	}

	expectedTitles := []string{"Moby Dick", "Clean Code", "iPhone"}
	for i, expected := range expectedTitles {
		if titles[i] != expected {
			t.Errorf("Expected title %s, got %s", expected, titles[i])
		}
	}
}

func TestParentNavigation(t *testing.T) {
	jsonData := []byte(`{
		"store": {
			"books": [
				{"title": "Moby Dick", "price": 8.99}
			]
		}
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Navigate to the first book's title
	bookTitle := root.Query("/store/books/0/title")
	if !bookTitle.IsValid() {
		t.Fatalf("Failed to navigate to book title: %v", bookTitle.Error())
	}

	// Navigate back to the parent (the book object)
	parent := bookTitle.Query("../")
	if !parent.IsValid() {
		t.Fatalf("Failed to navigate to parent: %v", parent.Error())
	}

	if parent.Type() != core.Object {
		t.Fatalf("Expected parent to be an object, got %v", parent.Type())
	}

	// Navigate back to the grandparent (the books array)
	grandparent := bookTitle.Query("../..")
	if !grandparent.IsValid() {
		t.Fatalf("Failed to navigate to grandparent: %v", grandparent.Error())
	}

	if grandparent.Type() != core.Array {
		t.Fatalf("Expected grandparent to be an array, got %v", grandparent.Type())
	}
}

func TestCombinedRecursiveAndParent(t *testing.T) {
	jsonData := []byte(`{
		"level1": {
			"level2": {
				"target": "found me"
			},
			"another": {
				"target": "found me too"
			}
		}
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Find all "target" fields recursively
	targets := root.Query("//target")
	if !targets.IsValid() {
		t.Fatalf("Recursive query failed: %v", targets.Error())
	}

	if targets.Type() != core.Array {
		t.Fatalf("Expected array result, got %v", targets.Type())
	}

	// Navigate from targets back to their parents
	results := targets.Array()
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Check first result
	firstParent := results[0].Query("../")
	if !firstParent.IsValid() {
		t.Fatalf("Failed to get parent of first result: %v", firstParent.Error())
	}

	// Check second result
	secondParent := results[1].Query("../")
	if !secondParent.IsValid() {
		t.Fatalf("Failed to get parent of second result: %v", secondParent.Error())
	}
}