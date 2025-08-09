package engine

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
)

func TestBaseNodeForEachAdditional(t *testing.T) {
	// Test baseNode ForEach which has 0% coverage
	bn := &baseNode{}
	var called bool
	bn.ForEach(func(key interface{}, value core.Node) {
		called = true
	})
	if called {
		t.Error("baseNode ForEach should not call the iterator")
	}
}

func TestArrayNodeRawMethod(t *testing.T) {
	// Test arrayNode Raw method which has 22.2% coverage
	// Test normal case
	node, err := ParseJSONToNode(`[1, 2, 3]`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if node.Error() != nil {
		t.Fatalf("node has an error: %v", node.Error())
	}
	arrNode := node.(*arrayNode)
	// Note: JSON marshaling adds spaces after commas
	if arrNode.Raw() != "[1, 2, 3]" {
		t.Errorf("expected [1, 2, 3], got %s", arrNode.Raw())
	}

	// Test with raw value set
	raw := "[4,5,6]"
	arrNodeWithRaw := &arrayNode{baseNode: baseNode{raw: &raw}}
	if arrNodeWithRaw.Raw() != "[4,5,6]" {
		t.Errorf("expected [4,5,6], got %s", arrNodeWithRaw.Raw())
	}

	// Test with error
	errNode := &arrayNode{baseNode: baseNode{err: errors.New("test error")}}
	if errNode.Raw() != "" {
		t.Errorf("expected empty string, got %s", errNode.Raw())
	}

	// Test with marshaling error - need to create a node that can't be marshaled
	// This is difficult to simulate, so we'll skip for now
}

func TestObjectNodeRawMethod(t *testing.T) {
	// Test objectNode Raw method which has 22.2% coverage
	// Test normal case
	node, err := ParseJSONToNode(`{"a": 1, "b": 2}`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if node.Error() != nil {
		t.Fatalf("node has an error: %v", node.Error())
	}
	objNode := node.(*objectNode)
	// Note: JSON marshaling adds spaces after colons
	if objNode.Raw() != `{"a": 1, "b": 2}` {
		t.Errorf("expected {\"a\": 1, \"b\": 2}, got %s", objNode.Raw())
	}

	// Test with raw value set
	raw := `{"c":3,"d":4}`
	objNodeWithRaw := &objectNode{baseNode: baseNode{raw: &raw}}
	if objNodeWithRaw.Raw() != `{"c":3,"d":4}` {
		t.Errorf("expected {\"c\":3,\"d\":4}, got %s", objNodeWithRaw.Raw())
	}

	// Test with error
	errNode := &objectNode{baseNode: baseNode{err: errors.New("test error")}}
	if errNode.Raw() != "" {
		t.Errorf("expected empty string, got %s", errNode.Raw())
	}
}

func TestApplyMethodCoverage(t *testing.T) {
	// Test arrayNode Apply method with different function types to improve coverage
	node, err := ParseJSONToNode(`[1, 2, 3, 4, 5]`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if node.Error() != nil {
		t.Fatalf("node has an error: %v", node.Error())
	}
	arrNode := node.(*arrayNode)

	// Test with PredicateFunc (Filter)
	predicateResult := arrNode.Apply(core.PredicateFunc(func(n core.Node) bool {
		return n.Int() > 2
	}))
	if predicateResult == nil {
		t.Fatal("predicateResult is nil")
	}
	if predicateResult.Type() != core.ArrayNode {
		t.Errorf("expected ArrayNode, got %v", predicateResult.Type())
	}

	// Check that filtered array has 3 elements [3, 4, 5]
	filteredArray := predicateResult.Array()
	if len(filteredArray) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(filteredArray))
	}
	if filteredArray[0].Int() != 3 {
		t.Errorf("expected 3, got %d", filteredArray[0].Int())
	}
	if filteredArray[1].Int() != 4 {
		t.Errorf("expected 4, got %d", filteredArray[1].Int())
	}
	if filteredArray[2].Int() != 5 {
		t.Errorf("expected 5, got %d", filteredArray[2].Int())
	}

	// Test with TransformFunc (Map)
	transformResult := arrNode.Apply(core.TransformFunc(func(n core.Node) interface{} {
		return n.Int() * 2
	}))
	if transformResult == nil {
		t.Fatal("transformResult is nil")
	}
	if transformResult.Type() != core.ArrayNode {
		t.Errorf("expected ArrayNode, got %v", transformResult.Type())
	}

	// Check that mapped array has doubled values [2, 4, 6, 8, 10]
	mappedArray := transformResult.Array()
	if len(mappedArray) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(mappedArray))
	}
	if mappedArray[0].Int() != 2 {
		t.Errorf("expected 2, got %d", mappedArray[0].Int())
	}
	if mappedArray[1].Int() != 4 {
		t.Errorf("expected 4, got %d", mappedArray[1].Int())
	}
	if mappedArray[2].Int() != 6 {
		t.Errorf("expected 6, got %d", mappedArray[2].Int())
	}
	if mappedArray[3].Int() != 8 {
		t.Errorf("expected 8, got %d", mappedArray[3].Int())
	}
	if mappedArray[4].Int() != 10 {
		t.Errorf("expected 10, got %d", mappedArray[4].Int())
	}

	// Test with unsupported function type
	unsupportedResult := arrNode.Apply(func() {})
	if unsupportedResult.Type() != core.InvalidNode {
		t.Errorf("expected InvalidNode, got %v", unsupportedResult.Type())
	}
	if unsupportedResult.Error() == nil || !strings.Contains(unsupportedResult.Error().Error(), "unsupported function signature") {
		t.Errorf("expected error containing 'unsupported function signature', got %v", unsupportedResult.Error())
	}

	// Test object Apply with PredicateFunc (Filter)
	objNode, err := ParseJSONToNode(`{"a": 1, "b": 2, "c": 3}`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if objNode.Error() != nil {
		t.Fatalf("node has an error: %v", objNode.Error())
	}
	objNodeCast := objNode.(*objectNode)

	objPredicateResult := objNodeCast.Apply(core.PredicateFunc(func(n core.Node) bool {
		return n.Int() > 1
	}))
	if objPredicateResult.Error() != nil {
		t.Errorf("expected no error, but got %v", objPredicateResult.Error())
	}

	// Test object Apply with TransformFunc (Map)
	objTransformResult := objNodeCast.Apply(core.TransformFunc(func(n core.Node) interface{} {
		return n.Int() * 10
	}))
	if objTransformResult.Error() != nil {
		t.Errorf("expected no error, but got %v", objTransformResult.Error())
	}

	// Test Apply on node with error
	errNode := &arrayNode{baseNode: baseNode{err: errors.New("test error")}}
	result := errNode.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
	if result != errNode {
		t.Errorf("expected same node, got different")
	}
}

func TestStringNodeMethods(t *testing.T) {
	// Test stringNode methods that have 66.7% coverage
	node, err := ParseJSONToNode(`"hello world"`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if node.Error() != nil {
		t.Fatalf("node has an error: %v", node.Error())
	}
	strNode := node.(*stringNode)

	// Test String method
	if strNode.String() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strNode.String())
	}

	// Test MustString method
	if strNode.MustString() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strNode.MustString())
	}

	// Test Interface method
	if strNode.Interface() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strNode.Interface())
	}

	// Test Time method with non-time string
	if strNode.Time() != (time.Time{}) {
		t.Errorf("expected zero time, got %v", strNode.Time())
	}
	// Reset error for subsequent tests
	strNode.err = nil

	// Test Raw method
	if strNode.Raw() != `"hello world"` {
		t.Errorf("expected '\"hello world\"', got '%s'", strNode.Raw())
	}

	// Test RawFloat method
	if _, ok := strNode.RawFloat(); ok {
		t.Error("expected RawFloat to fail")
	}

	// Test RawString method
	val, ok := strNode.RawString()
	if !ok {
		t.Error("expected RawString to succeed")
	}
	if val != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", val)
	}

	// Test Strings method
	strs := strNode.Strings()
	if len(strs) != 1 {
		t.Fatalf("expected 1 string, got %d", len(strs))
	}
	if strs[0] != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", strs[0])
	}

	// Test Contains method
	if !strNode.Contains("hello") {
		t.Error("expected Contains to return true for 'hello'")
	}
	if !strNode.Contains("world") {
		t.Error("expected Contains to return true for 'world'")
	}
	if strNode.Contains("universe") {
		t.Error("expected Contains to return false for 'universe'")
	}

	// Test Apply method with unsupported function
	result := strNode.Apply(func() {})
	if result.Type() != core.InvalidNode {
		t.Errorf("expected InvalidNode, got %v", result.Type())
	}
	if result.Error() == nil || !strings.Contains(result.Error().Error(), "unsupported function signature") {
		t.Errorf("expected error containing 'unsupported function signature', got %v", result.Error())
	}

	// Test Apply on node with error
	errNode := &stringNode{baseNode: baseNode{err: errors.New("test error")}}
	result = errNode.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
	if result != errNode {
		t.Error("expected same node, got different")
	}
}

func TestNumberNodeMethods(t *testing.T) {
	// Test numberNode methods that have 66.7% coverage
	node, err := ParseJSONToNode(`42.5`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if node.Error() != nil {
		t.Fatalf("node has an error: %v", node.Error())
	}
	numNode := node.(*numberNode)

	// Test String method
	if numNode.String() != "42.5" {
		t.Errorf("expected '42.5', got '%s'", numNode.String())
	}

	// Test Float and MustFloat methods
	if numNode.Float() != 42.5 {
		t.Errorf("expected 42.5, got %f", numNode.Float())
	}
	if numNode.MustFloat() != 42.5 {
		t.Errorf("expected 42.5, got %f", numNode.MustFloat())
	}

	// Test Int and MustInt methods
	if numNode.Int() != 42 {
		t.Errorf("expected 42, got %d", numNode.Int())
	}
	if numNode.MustInt() != 42 {
		t.Errorf("expected 42, got %d", numNode.MustInt())
	}

	// Test Interface method
	if numNode.Interface() != 42.5 {
		t.Errorf("expected 42.5, got %v", numNode.Interface())
	}

	// Test Raw method
	if numNode.Raw() != "42.5" {
		t.Errorf("expected '42.5', got '%s'", numNode.Raw())
	}

	// Test RawFloat method
	val, ok := numNode.RawFloat()
	if !ok {
		t.Error("expected RawFloat to succeed")
	}
	if val != 42.5 {
		t.Errorf("expected 42.5, got %f", val)
	}

	// Test Apply method with unsupported function
	result := numNode.Apply(func() {})
	if result.Type() != core.InvalidNode {
		t.Errorf("expected InvalidNode, got %v", result.Type())
	}
	if result.Error() == nil || !strings.Contains(result.Error().Error(), "unsupported function signature") {
		t.Errorf("expected error containing 'unsupported function signature', got %v", result.Error())
	}

	// Test Apply on node with error
	errNode := &numberNode{baseNode: baseNode{err: errors.New("test error")}}
	result = errNode.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
	if result != errNode {
		t.Error("expected same node, got different")
	}
}

func TestBoolNodeMethods(t *testing.T) {
	// Test boolNode methods that have 66.7% coverage
	node, err := ParseJSONToNode(`true`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if node.Error() != nil {
		t.Fatalf("node has an error: %v", node.Error())
	}
	boolNodeInstance := node.(*boolNode)

	// Test Bool and MustBool methods
	if !boolNodeInstance.Bool() {
		t.Error("expected true, got false")
	}
	if !boolNodeInstance.MustBool() {
		t.Error("expected true, got false")
	}

	// Test Interface method
	if !boolNodeInstance.Interface().(bool) {
		t.Error("expected true, got false")
	}

	// Test Raw method
	if boolNodeInstance.Raw() != "true" {
		t.Errorf("expected 'true', got '%s'", boolNodeInstance.Raw())
	}

	// Test Apply method with unsupported function
	result := boolNodeInstance.Apply(func() {})
	if result.Type() != core.InvalidNode {
		t.Errorf("expected InvalidNode, got %v", result.Type())
	}
	if result.Error() == nil || !strings.Contains(result.Error().Error(), "unsupported function signature") {
		t.Errorf("expected error containing 'unsupported function signature', got %v", result.Error())
	}

	// Test Apply on node with error
	errNode := &boolNode{baseNode: baseNode{err: errors.New("test error")}}
	result = errNode.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
	if result != errNode {
		t.Error("expected same node, got different")
	}
}

func TestNullNodeMethods(t *testing.T) {
	// Test nullNode methods that have low coverage
	node, err := ParseJSONToNode(`null`)
	if err != nil {
		t.Fatalf("ParseJSONToNode failed: %v", err)
	}
	if node.Error() != nil {
		t.Fatalf("node has an error: %v", node.Error())
	}
	nullNodeInstance := node.(*nullNode)

	// Test Raw method
	if nullNodeInstance.Raw() != "null" {
		t.Errorf("expected 'null', got '%s'", nullNodeInstance.Raw())
	}

	// Test Apply method with unsupported function
	result := nullNodeInstance.Apply(func() {})
	if result.Type() != core.InvalidNode {
		t.Errorf("expected InvalidNode, got %v", result.Type())
	}
	if result.Error() == nil || !strings.Contains(result.Error().Error(), "unsupported function signature") {
		t.Errorf("expected error containing 'unsupported function signature', got %v", result.Error())
	}

	// Test Apply on node with error
	errNode := &nullNode{baseNode: baseNode{err: errors.New("test error")}}
	result = errNode.Apply(core.PredicateFunc(func(n core.Node) bool { return true }))
	if result != errNode {
		t.Error("expected same node, got different")
	}
}
