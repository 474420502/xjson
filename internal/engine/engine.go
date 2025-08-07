// Package engine provides the query execution engine for xjson.
// It implements the core logic for executing parsed XPath-like queries
// on JSON data, supporting both raw bytes and materialized structures.
package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/474420502/xjson/internal/filter"
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
				// Use the new FilterEvaluator
				evaluator := filter.NewFilterEvaluator()
				for _, item := range arr {
					ctx := &filter.EvaluationContext{
						CurrentItem: item,
						RootData:    ctx.data,
					}
					match, err := evaluator.EvaluateExpression(pred.Expression, ctx)
					if err == nil && match {
						nextResults = append(nextResults, item)
					}
				}
			}
		}
		results = nextResults
	}
	return results
}

// LEGACY/DEPRECATED FUNCTIONS
func (e *Engine) ExecuteOnRaw(data []byte, query *parser.Query) ([]Match, error) {
	return nil, errors.New("ExecuteOnRaw is deprecated")
}

type QueryContext struct {
	scanner *scanner.Scanner
	data    []byte
}

func (e *Engine) getRootData(ctx *QueryContext) interface{} {
	return nil
}

func ApplyFilter(items []interface{}, predicate parser.Predicate, rootData interface{}) ([]interface{}, error) {
	return nil, errors.New("ApplyFilter is deprecated")
}

func ParseSimplePath(path string) []string {
	return strings.Split(path, ".")
}

func ConvertValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	return nil, errors.New("ConvertValue is deprecated")
}

func GetValueBySimplePathFromRaw(data []byte, path string) (interface{}, bool) {
	return nil, false
}

func parseJSONValue(data []byte) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal(data, &v)
	return v, err
}
