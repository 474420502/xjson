package engine

import (
	"reflect"
	"strings"
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
	expectedTitles := map[string]bool{"Moby Dick": true, "Clean Code": true, "iPhone": true}

	if len(titles) != len(expectedTitles) {
		t.Fatalf("Expected %d titles, got %d: %v", len(expectedTitles), len(titles), titles)
	}

	for _, title := range titles {
		if !expectedTitles[title] {
			t.Errorf("Unexpected title %s found", title)
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
	if parent.Get("price").Float() != 8.99 {
		t.Errorf("Parent object content is incorrect. Expected price 8.99, got %v", parent.Get("price").Float())
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

	parent1 := results[0].Query("../")
	parent2 := results[1].Query("../")

	if !parent1.IsValid() || !parent2.IsValid() {
		t.Fatalf("Failed to get parent of one of the results")
	}

	parent1Target := parent1.Get("target").String()
	parent2Target := parent2.Get("target").String()

	expectedTargets := map[string]bool{"found me": true, "found me too": true}

	if !expectedTargets[parent1Target] {
		t.Errorf("Unexpected parent1 target: %s", parent1Target)
	}
	if !expectedTargets[parent2Target] {
		t.Errorf("Unexpected parent2 target: %s", parent2Target)
	}
	if parent1Target == parent2Target {
		t.Errorf("Parents are expected to have different target values, but both were %s", parent1Target)
	}
}

func TestWildcardQueries(t *testing.T) {
	jsonData := []byte(`{
		"store": {
			"books": {"title": "Book 1", "price": 10},
			"bikes": {"color": "red", "price": 100}
		},
		"data": [
			{"id": 1},
			{"id": 2}
		]
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Run("Wildcard on object", func(t *testing.T) {
		storeItems := root.Query("/store/*")
		if !storeItems.IsValid() {
			t.Fatalf("Wildcard query on object failed: %v", storeItems.Error())
		}
		if storeItems.Type() != core.Array {
			t.Fatalf("Expected array result, got %v", storeItems.Type())
		}
		if storeItems.Len() != 2 {
			t.Fatalf("Expected 2 items, got %d", storeItems.Len())
		}
	})

	t.Run("Wildcard on object with sub-query", func(t *testing.T) {
		result := root.Query("/store/*/price")
		if !result.IsValid() {
			t.Fatalf("Wildcard query failed: %v", result.Error())
		}
		var prices []float64
		for _, n := range result.Array() {
			prices = append(prices, n.Float())
		}

		if len(prices) != 2 {
			t.Fatalf("Expected 2 prices, got %d", len(prices))
		}

		expectedPrices := map[float64]bool{10: true, 100: true}
		for _, p := range prices {
			if !expectedPrices[p] {
				t.Errorf("Unexpected price: %f", p)
			}
			delete(expectedPrices, p)
		}
		if len(expectedPrices) != 0 {
			t.Errorf("Not all expected prices found")
		}
	})

	t.Run("Wildcard on array", func(t *testing.T) {
		dataItems := root.Query("/data/*")
		if !dataItems.IsValid() {
			t.Fatalf("Wildcard query on array failed: %v", dataItems.Error())
		}
		if dataItems.Type() != core.Array {
			t.Fatalf("Expected array result, got %v", dataItems.Type())
		}
		if dataItems.Len() != 2 {
			t.Fatalf("Expected 2 items from array wildcard, got %d", dataItems.Len())
		}
		if dataItems.Query("0/id").Int() != 1 {
			t.Errorf("Expected first item id to be 1")
		}
	})
}

func TestSliceQueries(t *testing.T) {
	jsonData := []byte(`{"items": [0, 1, 2, 3, 4, 5]}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	testCases := []struct {
		name     string
		path     string
		expected []int64
		isValid  bool
	}{
		{"Single Index", "/items[2]", []int64{2}, true},
		{"Negative Index", "/items[-1]", []int64{5}, true},
		{"Simple Slice", "/items[1:4]", []int64{1, 2, 3}, true},
		{"Slice to End", "/items[3:]", []int64{3, 4, 5}, true},
		{"Slice from Start", "/items[:3]", []int64{0, 1, 2}, true},
		{"Negative Slice to End", "/items[-3:]", []int64{3, 4, 5}, true},
		{"Full Slice", "/items[:]", []int64{0, 1, 2, 3, 4, 5}, true},
		{"Out of Bounds High", "/items[10]", nil, false},
		{"Out of Bounds Negative", "/items[-10]", nil, false},
		{"Invalid Slice Range", "/items[4:1]", []int64{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := root.Query(tc.path)
			if result.IsValid() != tc.isValid {
				t.Fatalf("Expected IsValid to be %v, but it was %v for path %s (error: %v)", tc.isValid, result.IsValid(), tc.path, result.Error())
			}

			if !tc.isValid {
				return
			}

			if !result.IsValid() {
				t.Fatalf("Query failed for path %s: %v", tc.path, result.Error())
			}

			var values []int64 = make([]int64, 0)
			if result.Type() == core.Array {
				arr := result.Array()
				for _, item := range arr {
					values = append(values, item.Int())
				}
			} else if result.Type() == core.Number {
				values = []int64{result.Int()}
			} else {
				if !(result.Type() == core.Array && result.Len() == 0) {
					t.Fatalf("Unexpected type %v", result.Type())
				}
			}

			if !reflect.DeepEqual(values, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, values)
			}
		})
	}
}

func TestSpecialKeyQueries(t *testing.T) {
	jsonData := []byte(`{
		"user.profile": {"name": "dot"},
		"/api/v1/users": {"name": "slash"},
		"key with spaces": {"name": "space"},
		"a\"key": {"name": "double_quote"},
		"a'key": {"name": "single_quote"}
	}`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	testCases := []struct {
		name         string
		path         string
		expectedName string
	}{
		{"Dot in key", `['user.profile']/name`, "dot"},
		{"Slash in key", `["/api/v1/users"]/name`, "slash"},
		{"Space in key", `['key with spaces']/name`, "space"},
		{"Double quote in key", `['a"key']/name`, "double_quote"},
		{"Single quote in key", `["a'key"]/name`, "single_quote"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := root.Query(tc.path)
			if err := result.Error(); err != nil {
				t.Fatalf("Query for path '%s' failed: %v", tc.path, err)
			}
			name := result.String()
			if name != tc.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tc.expectedName, name)
			}
		})
	}
}

func TestInvalidPathQueries(t *testing.T) {
	jsonData := []byte(`{"a": {"b": [1, 2]}, "c": 1}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	testCases := []string{
		"/a/x",    // Non-existent key
		"/c/d",    // Key access on a non-object
		"/a/b/x",  // Key access on an array
		"/a/b[5]", // Index out of bounds
		"/x",      // Non-existent root key
		"//d",     // Recursive descent for non-existent key, should return empty valid array
		"/[0]",    // Index access on object
		"/a/b[a]", // Invalid index in path string
		"/a/..b",  // Invalid parent navigation syntax
		"/a@func", // Invalid function call syntax
		"books[",  // Unmatched bracket
		"['key",   // Unmatched quote
	}

	for _, path := range testCases {
		t.Run(path, func(t *testing.T) {
			result := root.Query(path)
			// For `//d`, the result should be a valid but empty array
			if strings.HasPrefix(path, "//") {
				if !result.IsValid() {
					t.Errorf("Expected path '%s' to be valid, but got invalid", path)
				}
				if result.Type() != core.Array || result.Len() != 0 {
					t.Errorf("Expected valid empty array for path '%s', but got %v", path, result.Raw())
				}
				return
			}

			if result.IsValid() {
				t.Errorf("Expected path '%s' to be invalid, but got a valid result: %v", path, result.Raw())
			}
		})
	}
}

func TestPathFunctions(t *testing.T) {
	jsonData := []byte(`{
		"books": [
			{"title": "Moby Dick", "price": 8.99},
			{"title": "Clean Code", "price": 29.99},
			{"title": "The Hobbit", "price": 12.99}
		]
	}`)
	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	root.RegisterFunc("cheap", func(n core.Node) core.Node {
		return n.Filter(func(child core.Node) bool {
			price, ok := child.Get("price").RawFloat()
			return ok && price < 20
		})
	})

	root.RegisterFunc("expensive", func(n core.Node) core.Node {
		return n.Filter(func(child core.Node) bool {
			price, ok := child.Get("price").RawFloat()
			return ok && price >= 20
		})
	})

	t.Run("Valid Function Call", func(t *testing.T) {
		result := root.Query("/books[@cheap]/title")
		if !result.IsValid() {
			t.Fatalf("Query with function failed: %v", result.Error())
		}

		titles := result.Strings()
		expected := map[string]bool{"Moby Dick": true, "The Hobbit": true}

		if len(titles) != len(expected) {
			t.Fatalf("Expected %d titles, got %d. Got %v", len(expected), len(titles), titles)
		}
		for _, title := range titles {
			if !expected[title] {
				t.Errorf("Unexpected title '%s'", title)
			}
		}
	})

	t.Run("Chained Function Calls", func(t *testing.T) {
		result := root.Query("/books[@cheap][@expensive]")
		if !result.IsValid() {
			t.Fatalf("Chained function call query failed: %v", result.Error())
		}
		if result.Len() != 0 {
			t.Errorf("Expected 0 results for cheap and expensive filter, got %d", result.Len())
		}
	})

	t.Run("Non-existent Function", func(t *testing.T) {
		result := root.Query("/books[@nonexistent]")
		if result.IsValid() {
			t.Errorf("Expected invalid result for non-existent function, but was valid")
		}
	})

	t.Run("Function on non-array value", func(t *testing.T) {
		root.RegisterFunc("itself", func(n core.Node) core.Node {
			return n
		})
		result := root.Query("/books/0[@itself]/title")
		if !result.IsValid() {
			t.Fatalf("Function on non-array failed: %v", result.Error())
		}
		if result.String() != "Moby Dick" {
			t.Errorf("Expected 'Moby Dick', got '%s'", result.String())
		}
	})
}
