package xjson

import (
	"testing"
	"time"
)

func TestDottedKeyQuery(t *testing.T) {
	doc, _ := ParseString(`{"key.with.dots": "value", "normal": {"key.with.dots": "nested"}}`)

	// 设置超时
	done := make(chan bool, 1)

	go func() {
		result := doc.Query("key.with.dots")
		t.Logf("Query result exists: %v", result.Exists())
		done <- true
	}()

	select {
	case <-done:
		t.Log("Query completed successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("Query timed out - likely infinite loop")
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
