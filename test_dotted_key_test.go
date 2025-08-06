package xjson

import (
	"testing"
	"time"
)

func TestDottedKeyQuery(t *testing.T) {
	doc, _ := ParseString(`{"key.with.dots": "value", "normal": {"key.with.dots": "nested"}}`)

	// 1. Test with the old, deprecated dot notation.
	// This should still work but will print a warning.
	// We expect it to find the nested value, not the top-level one.
	resultDot := doc.Query("normal.key.with.dots")
	if !resultDot.Exists() {
		t.Errorf("Expected to find value with dot notation query, but it didn't exist.")
	}
	if val, _ := resultDot.String(); val != "nested" {
		t.Errorf("Expected 'nested', got '%s' for dot notation query", val)
	}

	// 2. Test with XPath-style query for a key containing dots.
	// The key "key.with.dots" should be treated as a single identifier.
	// A simple query "key.with.dots" should NOT work as it's not a valid XPath without a leading '/'.
	resultInvalidPath := doc.Query("key.with.dots")
	if resultInvalidPath.Exists() {
		t.Errorf("Query 'key.with.dots' should not exist without a proper path separator, but it does.")
	}

	// 3. The correct XPath way to query a key with dots is by treating it as a literal name.
	resultXPath := doc.Query("/key.with.dots")
	if !resultXPath.Exists() {
		t.Errorf("Expected to find value with XPath query for dotted key, but it didn't exist.")
	}
	if val, _ := resultXPath.String(); val != "value" {
		t.Errorf("Expected 'value', got '%s' for XPath query", val)
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
