// Package engine provides the query execution engine for xjson.
// It implements the core logic for executing parsed XPath-like queries
// on JSON data, supporting both raw bytes and materialized structures.
package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/474420502/xjson/internal/parser"
	"github.com/474420502/xjson/internal/scanner"
)

// Engine executes queries on JSON data
type Engine struct {
	scanner *scanner.Scanner
}

// NewEngine creates a new query execution engine
func NewEngine() *Engine {
	return &Engine{}
}

// ExecuteOnMaterialized executes a query on materialized Go structures
func (e *Engine) ExecuteOnMaterialized(data interface{}, query *parser.Query) ([]Match, error) {
	if query == nil || len(query.Steps) == 0 {
		// If there's no query, we match the entire document.
		return []Match{{Value: data}}, nil
	}
	context := &MaterializedContext{data: data}
	return e.executeStepsOnMaterialized(context, query.Steps, []interface{}{data})
}

// Match represents a query match result
type Match struct {
	Value interface{} // The matched value
}

// MaterializedContext holds the context for materialized query execution
type MaterializedContext struct {
	data interface{}
}

// executeStepsOnMaterialized recursively executes query steps on a list of current items
func (e *Engine) executeStepsOnMaterialized(ctx *MaterializedContext, steps []parser.Step, currentItems []interface{}) ([]Match, error) {
	if len(steps) == 0 {
		var matches []Match
		for _, item := range currentItems {
			matches = append(matches, Match{Value: item})
		}
		return matches, nil
	}

	step := steps[0]
	remaining := steps[1:]
	var nextItems []interface{}

	for _, current := range currentItems {
		var stepResults []interface{}
		switch step.Type {
		case parser.StepChild:
			stepResults = e.executeChildStepOnMaterialized(ctx, step, current)
		case parser.StepDescendant:
			stepResults = e.executeDescendantStepOnMaterialized(ctx, step, current)
		case parser.StepWildcard:
			stepResults = e.executeWildcardStepOnMaterialized(ctx, step, current)
		default:
			return nil, fmt.Errorf("unsupported step type: %v", step.Type)
		}
		nextItems = append(nextItems, stepResults...)
	}

	return e.executeStepsOnMaterialized(ctx, remaining, nextItems)
}

// executeChildStepOnMaterialized executes a child access step
func (e *Engine) executeChildStepOnMaterialized(ctx *MaterializedContext, step parser.Step, current interface{}) []interface{} {
	var results []interface{}
	if obj, ok := current.(map[string]interface{}); ok {
		if val, exists := obj[step.Name]; exists {
			results = append(results, val)
		}
	} else if step.Name == "" { // Allows predicates on the root if it's an array
		results = append(results, current)
	}

	if len(step.Predicates) > 0 {
		return e.applyPredicates(ctx, results, step.Predicates)
	}
	return results
}

// executeDescendantStepOnMaterialized executes a descendant step (e.g., //)
func (e *Engine) executeDescendantStepOnMaterialized(ctx *MaterializedContext, step parser.Step, current interface{}) []interface{} {
	var matches []interface{}
	var search func(interface{})

	search = func(node interface{}) {
		if m, ok := node.(map[string]interface{}); ok {
			if val, exists := m[step.Name]; exists {
				matches = append(matches, val)
			}
			for _, v := range m {
				search(v)
			}
		} else if a, ok := node.([]interface{}); ok {
			for _, v := range a {
				search(v)
			}
		}
	}

	search(current)
	return matches
}

// executeWildcardStepOnMaterialized executes a wildcard step (*)
func (e *Engine) executeWildcardStepOnMaterialized(ctx *MaterializedContext, step parser.Step, current interface{}) []interface{} {
	var results []interface{}
	if obj, ok := current.(map[string]interface{}); ok {
		for _, v := range obj {
			results = append(results, v)
		}
	} else if arr, ok := current.([]interface{}); ok {
		results = append(results, arr...)
	}
	return results
}

// applyPredicates handles predicates like [0], [-1], [1:3], [?(@.price > 10)]
func (e *Engine) applyPredicates(ctx *MaterializedContext, inputs []interface{}, predicates []parser.Predicate) []interface{} {
	results := inputs
	for _, pred := range predicates {
		var nextResults []interface{}
		for _, input := range results {
			arr, ok := input.([]interface{})
			if !ok {
				continue
			}

			switch pred.Type {
			case parser.PredicateIndex:
				idx := pred.Index
				if idx < 0 {
					idx += len(arr)
				}
				if idx >= 0 && idx < len(arr) {
					nextResults = append(nextResults, arr[idx])
				}
			case parser.PredicateSlice:
				start, end := pred.Start, pred.End
				if start < 0 {
					start += len(arr)
				}
				if end < 0 {
					end += len(arr)
				}
				if start < 0 {
					start = 0
				}
				if end > len(arr) {
					end = len(arr)
				}
				if start < end {
					nextResults = append(nextResults, arr[start:end]...)
				}
			case parser.PredicateExpression:
				filtered, _ := ApplyFilter(arr, pred, ctx.data)
				for _, item := range filtered {
					nextResults = append(nextResults, item)
				}
			}
		}
		results = nextResults
	}
	return results
}

// LEGACY/DEPRECATED FUNCTIONS BELOW
// These functions support the old dot-notation path and will be removed.

func ApplyFilter(items []interface{}, predicate parser.Predicate, rootData interface{}) (map[int]interface{}, error) {
	if predicate.Type != parser.PredicateExpression {
		return nil, errors.New("not an expression predicate")
	}
	results := make(map[int]interface{})
	for i, item := range items {
		match, err := evaluateFilterExpression(predicate.Expression, item, rootData)
		if err != nil {
			// In a real scenario, you might want to handle errors differently
			continue
		}
		if match {
			results[i] = item
		}
	}
	return results, nil
}

func evaluateFilterExpression(expr parser.Expression, currentItem, rootData interface{}) (bool, error) {
	switch expr.Type {
	case parser.ExpressionBinary:
		return evaluateBinaryExpression(expr, currentItem, rootData)
	case parser.ExpressionLiteral:
		return toBooleanValue(expr.Value), nil
	case parser.ExpressionPath:
		val, err := getFilterExpressionValue(expr, currentItem, rootData)
		if err != nil {
			return false, nil
		}
		return toBooleanValue(val), nil
	default:
		return false, fmt.Errorf("unsupported expression type: %v", expr.Type)
	}
}

func evaluateBinaryExpression(expr parser.Expression, currentItem, rootData interface{}) (bool, error) {
	leftVal, err := getFilterExpressionValue(*expr.Left, currentItem, rootData)
	if err != nil {
		return false, err
	}

	if expr.Operator == "&&" {
		if !toBooleanValue(leftVal) {
			return false, nil
		}
		rightVal, err := getFilterExpressionValue(*expr.Right, currentItem, rootData)
		if err != nil {
			return false, err
		}
		return toBooleanValue(rightVal), nil
	}
	if expr.Operator == "||" {
		if toBooleanValue(leftVal) {
			return true, nil
		}
		rightVal, err := getFilterExpressionValue(*expr.Right, currentItem, rootData)
		if err != nil {
			return false, err
		}
		return toBooleanValue(rightVal), nil
	}

	rightVal, err := getFilterExpressionValue(*expr.Right, currentItem, rootData)
	if err != nil {
		return false, err
	}
	return compareFilterValues(leftVal, rightVal, expr.Operator)
}

func getFilterExpressionValue(expr parser.Expression, currentItem, rootData interface{}) (interface{}, error) {
	switch expr.Type {
	case parser.ExpressionLiteral:
		return expr.Value, nil
	case parser.ExpressionPath:
		path := expr.Path
		var startNode interface{}
		if len(path) > 0 {
			if path[0] == "@" {
				startNode = currentItem
				path = path[1:]
			} else if path[0] == "$" {
				startNode = rootData
				path = path[1:]
			}
		}
		val, found := GetValueBySimplePath(startNode, strings.Join(path, "."))
		if !found {
			return nil, errors.New("path not found")
		}
		return val, nil
	default:
		return nil, fmt.Errorf("unsupported expression type for value retrieval: %v", expr.Type)
	}
}

func compareFilterValues(left, right interface{}, operator string) (bool, error) {
	if left == nil || right == nil {
		if operator == "==" {
			return left == right, nil
		}
		if operator == "!=" {
			return left != right, nil
		}
		return false, nil
	}

	leftFloat, leftIsNum := convertToFloat(left)
	rightFloat, rightIsNum := convertToFloat(right)

	if leftIsNum && rightIsNum {
		switch operator {
		case "==":
			return leftFloat == rightFloat, nil
		case "!=":
			return leftFloat != rightFloat, nil
		case ">":
			return leftFloat > rightFloat, nil
		case ">=":
			return leftFloat >= rightFloat, nil
		case "<":
			return leftFloat < rightFloat, nil
		case "<=":
			return leftFloat <= rightFloat, nil
		}
	}

	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	if operator == "==" {
		return leftStr == rightStr, nil
	}
	if operator == "!=" {
		return leftStr != rightStr, nil
	}

	return false, fmt.Errorf("unsupported operator %s for types %T and %T", operator, left, right)
}

func convertToFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	}
	return 0, false
}

func toBooleanValue(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != ""
	case float64:
		return val != 0
	case int, int64:
		return val != 0
	case nil:
		return false
	default:
		return true
	}
}

func GetValueBySimplePath(data interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		if part == "" {
			continue
		}
		if obj, ok := current.(map[string]interface{}); ok {
			if val, exists := obj[part]; exists {
				current = val
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return current, true
}
