package xjson

import (
	"testing"
	"time"
)

func TestDottedKeyQuery(t *testing.T) {
	doc, _ := ParseString(`{"key.with.dots": "value", "normal": {"key.with.dots": "nested"}}`)

	// 1. Test that a query with dot notation is treated as a path, and does not match a literal key with dots.
	resultAsPath := doc.Query(`/key/with/dots`)
	if resultAsPath.Exists() {
		t.Errorf("Query '/key.with.dots' should be treated as a path and not exist for a literal key, but it was found.")
	}

	// 2. The correct way to query a key with dots is by quoting it.
	resultXPath := doc.Query(`/"key.with.dots"`)
	if !resultXPath.Exists() {
		t.Errorf("Expected to find value with quoted XPath query for dotted key, but it didn't exist.")
	}
	if val, _ := resultXPath.String(); val != "value" {
		t.Errorf("Expected 'value', got '%s' for quoted XPath query", val)
	}

	// 3. The correct way to query a nested key with dots is also by quoting it.
	resultNestedXPath := doc.Query(`/normal/"key.with.dots"`)
	if !resultNestedXPath.Exists() {
		t.Errorf("Expected to find value with quoted XPath query for nested dotted key, but it didn't exist.")
	}
	if val, _ := resultNestedXPath.String(); val != "nested" {
		t.Errorf("Expected 'nested', got '%s' for nested XPath query", val)
	}

	// 4. A query without a root path separator should still be invalid.
	resultInvalidPath := doc.Query(`key.with.dots`)
	if resultInvalidPath.Exists() {
		t.Errorf("Query 'key.with.dots' should not exist without a proper path separator, but it does.")
	}
}

func TestEmptyStringQuery(t *testing.T) {
	doc, _ := ParseString(`{"key": "value"}`)

	// 设置超时
	done := make(chan bool, 1)

	go func() {
		result := doc.Query("")
		t.Logf("Empty query result exists: %v", result.Exists())
		done <- true
	}()

	select {
	case <-done:
		t.Log("Empty query completed successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("Empty query timed out - likely infinite loop")
	}
}

func TestRootQuery(t *testing.T) {
	doc, _ := ParseString(`{"key": "value"}`)

	// 设置超时
	done := make(chan bool, 1)

	go func() {
		result := doc.Query("$")
		t.Logf("Root query result exists: %v", result.Exists())
		done <- true
	}()

	select {
	case <-done:
		t.Log("Root query completed successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("Root query timed out - likely infinite loop")
	}
}
