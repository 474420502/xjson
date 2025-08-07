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
	var stepResults []interface{}

	if step.Type == parser.StepDescendant {
		stepResults = e.executeDescendantStepOnMaterialized(ctx, step, currentItems[0])
	} else {
		for _, current := range currentItems {
			var items []interface{}
			switch step.Type {
			case parser.StepChild:
				items = e.executeChildStepOnMaterialized(ctx, step, current)
			case parser.StepWildcard:
				items = e.executeWildcardStepOnMaterialized(ctx, step, current)
			default:
				return nil, fmt.Errorf("unsupported step type: %v", step.Type)
			}
			stepResults = append(stepResults, items...)
		}
	}

	if len(step.Predicates) > 0 {
		stepResults = e.applyPredicates(ctx, stepResults, step.Predicates)
	}

	return e.executeStepsOnMaterialized(ctx, remaining, stepResults)
}

// executeChildStepOnMaterialized executes a child access step
func (e *Engine) executeChildStepOnMaterialized(ctx *MaterializedContext, step parser.Step, current interface{}) []interface{} {
	if obj, ok := current.(map[string]interface{}); ok {
		if val, exists := obj[step.Name]; exists {
			if arr, isArr := val.([]interface{}); isArr {
				return arr
			}
			return []interface{}{val}
		}
	} else if arr, ok := current.([]interface{}); ok {
		return arr
	}
	return nil
}

// executeDescendantStepOnMaterialized executes a descendant step (e.g., //)
func (e *Engine) executeDescendantStepOnMaterialized(ctx *MaterializedContext, step parser.Step, current interface{}) []interface{} {
	var matches []interface{}
	var search func(node interface{})

	search = func(node interface{}) {
		if m, ok := node.(map[string]interface{}); ok {
			if val, exists := m[step.Name]; exists {
				if arr, isArr := val.([]interface{}); isArr {
					matches = append(matches, arr...)
				} else {
					matches = append(matches, val)
				}
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

// applyPredicates handles predicates like [1], [@key='value']
func (e *Engine) applyPredicates(ctx *MaterializedContext, itemsToFilter []interface{}, predicates []parser.Predicate) []interface{} {
	results := itemsToFilter
	for _, pred := range predicates {
		var nextResults []interface{}

		switch pred.Type {
		case parser.PredicateIndex:
			idx := pred.Index
			if idx < 0 {
				idx += len(results)
			}
			if idx >= 0 && idx < len(results) {
				nextResults = append(nextResults, results[idx])
			}

		case parser.PredicateExpression:
			evaluator := filter.NewFilterEvaluator()
			for i, item := range results {
				evalCtx := &filter.EvaluationContext{
					ContextNode: item,
					RootData:    ctx.data,
					Position:    i + 1,
					Size:        len(results),
				}
				match, err := evaluator.EvaluateExpression(pred.Expression, evalCtx)
				if err == nil && match {
					nextResults = append(nextResults, item)
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
