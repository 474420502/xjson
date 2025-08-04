// Package engine provides the query execution engine for xjson.
// It implements the core logic for executing parsed XPath-like queries
// on JSON data, supporting both raw bytes and materialized structures.
package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
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

// ExecuteOnRaw executes a query on raw JSON bytes
func (e *Engine) ExecuteOnRaw(data []byte, query *parser.Query) ([]Match, error) {
	e.scanner = scanner.NewScanner(data)

	var matches []Match

	// For now, implement simple path-based queries
	if len(query.Steps) == 0 {
		return matches, nil
	}

	// Start from root
	context := &QueryContext{
		scanner: e.scanner,
		data:    data,
	}

	result, err := e.executeSteps(context, query.Steps, 0)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteOnMaterialized executes a query on materialized Go structures
func (e *Engine) ExecuteOnMaterialized(data interface{}, query *parser.Query) ([]Match, error) {
	var matches []Match

	// For now, implement simple path-based queries
	if len(query.Steps) == 0 {
		return matches, nil
	}

	context := &MaterializedContext{
		data: data,
	}

	result, err := e.executeStepsOnMaterialized(context, query.Steps, data)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Match represents a query match result
type Match struct {
	Value interface{} // The matched value
	Path  string      // The path to the matched value
	Raw   []byte      // Raw bytes if from raw scanning
}

// QueryContext holds the context for raw query execution
type QueryContext struct {
	scanner *scanner.Scanner
	data    []byte
}

// MaterializedContext holds the context for materialized query execution
type MaterializedContext struct {
	data interface{}
}

// executeSteps executes query steps on raw JSON data
func (e *Engine) executeSteps(ctx *QueryContext, steps []parser.Step, startPos int) ([]Match, error) {
	var matches []Match

	if len(steps) == 0 {
		// Return the current value
		if value, ok := ctx.scanner.GetValueAt(); ok {
			match := Match{
				Raw:  value,
				Path: "",
			}
			// Try to parse the value
			var parsed interface{}
			if err := json.Unmarshal(value, &parsed); err == nil {
				match.Value = parsed
			}
			matches = append(matches, match)
		}
		return matches, nil
	}

	step := steps[0]
	remaining := steps[1:]

	switch step.Type {
	case parser.StepChild:
		return e.executeChildStep(ctx, step, remaining)
	case parser.StepDescendant:
		return e.executeDescendantStep(ctx, step, remaining)
	case parser.StepWildcard:
		return e.executeWildcardStep(ctx, step, remaining)
	default:
		return matches, fmt.Errorf("unsupported step type: %v", step.Type)
	}
}

// executeChildStep executes a child step (direct child access)
func (e *Engine) executeChildStep(ctx *QueryContext, step parser.Step, remaining []parser.Step) ([]Match, error) {
	var matches []Match

	ctx.scanner.SkipWhitespace()

	if ctx.scanner.Current() == '{' {
		// Object access
		if step.Name != "" {
			if ctx.scanner.FindKey(step.Name) {
				// Check if this step has predicates
				if len(step.Predicates) > 0 {
					return e.executeArrayStep(ctx, step, remaining)
				}

				if len(remaining) == 0 {
					// This is the final step, get the value
					if value, ok := ctx.scanner.GetValueAt(); ok {
						match := Match{
							Raw:  value,
							Path: step.Name,
						}
						var parsed interface{}
						if err := json.Unmarshal(value, &parsed); err == nil {
							match.Value = parsed
						}
						matches = append(matches, match)
					}
				} else {
					// Continue with remaining steps
					subMatches, err := e.executeSteps(ctx, remaining, ctx.scanner.Position())
					if err != nil {
						return nil, err
					}
					for _, match := range subMatches {
						match.Path = step.Name + "." + match.Path
						matches = append(matches, match)
					}
				}
			}
		}
	} else if ctx.scanner.Current() == '[' {
		// Array access - handle predicates
		return e.executeArrayStep(ctx, step, remaining)
	}

	return matches, nil
}

// executeDescendantStep executes a descendant step (recursive search)
func (e *Engine) executeDescendantStep(ctx *QueryContext, step parser.Step, remaining []parser.Step) ([]Match, error) {
	var matches []Match

	// 如果step有具体的名字，我们需要在所有深度上查找这个名字
	if step.Name != "" {
		// 递归搜索所有层级中名为step.Name的字段
		value, ok := ctx.scanner.GetValueAt()
		if !ok {
			return matches, nil
		}

		var current interface{}
		if err := json.Unmarshal(value, &current); err != nil {
			return matches, nil
		}

		// 递归搜索函数
		var searchForField func(node interface{}, path string, depth int)
		searchForField = func(node interface{}, path string, depth int) {
			// 防止深度过深和循环引用
			if depth > 50 { // 减少最大深度
				return
			}

			switch v := node.(type) {
			case map[string]interface{}:
				// 检查当前对象是否有目标字段
				if value, exists := v[step.Name]; exists {
					if len(remaining) == 0 {
						// 这是最终步骤，返回字段值
						match := Match{
							Value: value,
							Path:  step.Name,
						}
						if path != "" {
							match.Path = path + "." + step.Name
						}
						if rawBytes, err := json.Marshal(value); err == nil {
							match.Raw = rawBytes
						}
						matches = append(matches, match)
					} else {
						// 继续处理剩余步骤
						fieldBytes, err := json.Marshal(value)
						if err == nil {
							fieldCtx := &QueryContext{
								scanner: scanner.NewScanner(fieldBytes),
								data:    fieldBytes,
							}
							subMatches, err := e.executeSteps(fieldCtx, remaining, 0)
							if err == nil {
								for _, subMatch := range subMatches {
									newPath := step.Name
									if subMatch.Path != "" {
										newPath = step.Name + "." + subMatch.Path
									}
									if path != "" {
										newPath = path + "." + newPath
									}
									subMatch.Path = newPath
									matches = append(matches, subMatch)
								}
							}
						}
					}
				}

				// 递归搜索所有子对象
				for key, child := range v {
					childPath := key
					if path != "" {
						childPath = path + "." + key
					}
					searchForField(child, childPath, depth+1)
				}
			case []interface{}:
				// 递归搜索数组中的每个元素
				for i, child := range v {
					childPath := fmt.Sprintf("[%d]", i)
					if path != "" {
						childPath = path + "." + fmt.Sprintf("[%d]", i)
					}
					searchForField(child, childPath, depth+1)
				}
			}
		}

		searchForField(current, "", 0)
	} else {
		// 没有具体名字的递归查询 - 避免这种情况下的复杂递归
		return matches, nil
	}

	return matches, nil
}

// executeWildcardStep executes a wildcard step
func (e *Engine) executeWildcardStep(ctx *QueryContext, step parser.Step, remaining []parser.Step) ([]Match, error) {
	matches := []Match{}

	// Reset scanner to current position
	ctx.scanner.Reset()
	ctx.scanner.SkipWhitespace()

	if ctx.scanner.IsEOF() {
		return matches, nil
	}

	current := ctx.scanner.Current()

	switch current {
	case '{':
		// Object - iterate through all keys
		ctx.scanner.Advance(1) // Skip '{'
		ctx.scanner.SkipWhitespace()

		for !ctx.scanner.IsEOF() && ctx.scanner.Current() != '}' {
			// Read key
			if ctx.scanner.Current() != '"' {
				break
			}

			key, ok := ctx.scanner.ReadString()
			if !ok {
				break
			}

			ctx.scanner.SkipWhitespace()
			if ctx.scanner.Current() != ':' {
				break
			}
			ctx.scanner.Advance(1) // Skip ':'
			ctx.scanner.SkipWhitespace()

			// Get value position
			valueStart := ctx.scanner.Position()
			if !ctx.scanner.SkipValue() {
				break
			}
			valueEnd := ctx.scanner.Position()

			// Create match for this key-value pair
			valueBytes := ctx.data[valueStart:valueEnd]
			value, _ := parseJSONValue(valueBytes)

			match := Match{
				Value: value,
				Path:  key,
			}

			// Apply remaining steps if any
			if len(remaining) > 0 {
				// Create new context for the value
				valueCtx := &QueryContext{
					scanner: scanner.NewScanner(valueBytes),
					data:    valueBytes,
				}
				subMatches, err := e.executeSteps(valueCtx, remaining, 0)
				if err != nil {
					return nil, err
				}
				matches = append(matches, subMatches...)
			} else {
				matches = append(matches, match)
			}

			// Skip comma if present
			ctx.scanner.SkipWhitespace()
			if ctx.scanner.Current() == ',' {
				ctx.scanner.Advance(1)
				ctx.scanner.SkipWhitespace()
			}
		}

	case '[':
		// Array - iterate through all elements
		ctx.scanner.Advance(1) // Skip '['
		ctx.scanner.SkipWhitespace()

		index := 0
		for !ctx.scanner.IsEOF() && ctx.scanner.Current() != ']' {
			// Get value position
			valueStart := ctx.scanner.Position()
			if !ctx.scanner.SkipValue() {
				break
			}
			valueEnd := ctx.scanner.Position()

			// Create match for this array element
			valueBytes := ctx.data[valueStart:valueEnd]
			value, _ := parseJSONValue(valueBytes)

			match := Match{
				Value: value,
				Path:  fmt.Sprintf("[%d]", index),
			}

			// Apply remaining steps if any
			if len(remaining) > 0 {
				// Create new context for the value
				valueCtx := &QueryContext{
					scanner: scanner.NewScanner(valueBytes),
					data:    valueBytes,
				}
				subMatches, err := e.executeSteps(valueCtx, remaining, 0)
				if err != nil {
					return nil, err
				}
				matches = append(matches, subMatches...)
			} else {
				matches = append(matches, match)
			}

			index++

			// Skip comma if present
			ctx.scanner.SkipWhitespace()
			if ctx.scanner.Current() == ',' {
				ctx.scanner.Advance(1)
				ctx.scanner.SkipWhitespace()
			}
		}
	}

	return matches, nil
}

// executeArrayStep executes array access with predicates
func (e *Engine) executeArrayStep(ctx *QueryContext, step parser.Step, remaining []parser.Step) ([]Match, error) {
	var matches []Match

	// Get the current array value
	value, ok := ctx.scanner.GetValueAt()
	if !ok {
		return matches, nil
	}

	var array []interface{}
	if err := json.Unmarshal(value, &array); err != nil {
		return matches, nil
	}

	// Process each predicate
	for _, predicate := range step.Predicates {
		switch predicate.Type {
		case parser.PredicateIndex:
			// Simple array index access like [0] or [-1]
			index := predicate.Index
			// Handle negative indices
			if index < 0 {
				index = len(array) + index
			}

			if index >= 0 && index < len(array) {
				if len(remaining) == 0 {
					// This is the final step, return the array element
					match := Match{
						Value: array[index],
						Path:  fmt.Sprintf("[%d]", predicate.Index),
					}
					if rawBytes, err := json.Marshal(array[index]); err == nil {
						match.Raw = rawBytes
					}
					matches = append(matches, match)
				} else {
					// Continue with remaining steps on the array element
					elementBytes, err := json.Marshal(array[index])
					if err != nil {
						continue
					}

					elementCtx := &QueryContext{
						scanner: scanner.NewScanner(elementBytes),
						data:    elementBytes,
					}

					subMatches, err := e.executeSteps(elementCtx, remaining, 0)
					if err == nil {
						for _, subMatch := range subMatches {
							subMatch.Path = fmt.Sprintf("[%d].%s", predicate.Index, subMatch.Path)
							matches = append(matches, subMatch)
						}
					}
				}
			}

		case parser.PredicateSlice:
			// Array slice access like [1:3]
			start := predicate.Start
			end := predicate.End

			// Handle negative indices
			if start < 0 {
				start = len(array) + start
			}
			if end < 0 {
				end = len(array) + end
			}

			// Ensure bounds
			if start < 0 {
				start = 0
			}
			if end > len(array) {
				end = len(array)
			}

			if start < end {
				if len(remaining) == 0 {
					// This is the final step, return the slice
					sliceArray := array[start:end]
					match := Match{
						Value: sliceArray,
						Path:  fmt.Sprintf("[%d:%d]", predicate.Start, predicate.End),
					}
					if rawBytes, err := json.Marshal(sliceArray); err == nil {
						match.Raw = rawBytes
					}
					matches = append(matches, match)
				} else {
					// Continue with remaining steps on each element in the slice
					for i := start; i < end; i++ {
						elementBytes, err := json.Marshal(array[i])
						if err != nil {
							continue
						}

						elementCtx := &QueryContext{
							scanner: scanner.NewScanner(elementBytes),
							data:    elementBytes,
						}

						subMatches, err := e.executeSteps(elementCtx, remaining, 0)
						if err == nil {
							for _, subMatch := range subMatches {
								subMatch.Path = fmt.Sprintf("[%d].%s", i, subMatch.Path)
								matches = append(matches, subMatch)
							}
						}
					}
				}
			}

		case parser.PredicateExpression:
			// Filter expression like [?(@.price < 20)]
			rootData := e.getRootData(ctx)

			filteredMap, err := ApplyFilter(array, predicate, rootData)
			if err != nil {
				return nil, fmt.Errorf("error applying filter: %w", err)
			}

			if len(remaining) == 0 {
				// This is the final step, return filtered elements
				for idx, item := range filteredMap {
					match := Match{
						Value: item,
						Path:  fmt.Sprintf("[%d]", idx),
					}
					if rawBytes, err := json.Marshal(item); err == nil {
						match.Raw = rawBytes
					}
					matches = append(matches, match)
				}
			} else {
				// Continue with remaining steps on each filtered element
				for idx, item := range filteredMap {
					elementBytes, err := json.Marshal(item)
					if err != nil {
						continue
					}

					elementCtx := &QueryContext{
						scanner: scanner.NewScanner(elementBytes),
						data:    elementBytes,
					}

					subMatches, err := e.executeSteps(elementCtx, remaining, 0)
					if err == nil {
						for _, subMatch := range subMatches {
							subMatch.Path = fmt.Sprintf("[%d].%s", idx, subMatch.Path)
							matches = append(matches, subMatch)
						}
					}
				}
			}
		}
	}

	return matches, nil
}

// getRootData attempts to get the root data for filter context
func (e *Engine) getRootData(ctx *QueryContext) interface{} {
	// Try to get the full document from scanner
	// We need to navigate back to the root of the document
	originalPos := ctx.scanner.Position()

	// Reset to beginning and parse the entire document
	ctx.scanner.Reset()
	if rootValue, ok := ctx.scanner.GetValueAt(); ok {
		var rootData interface{}
		if err := json.Unmarshal(rootValue, &rootData); err == nil {
			// Restore original position
			ctx.scanner.SetPosition(originalPos)
			return rootData
		}
	}

	// Restore original position
	ctx.scanner.SetPosition(originalPos)
	return nil
}

// executeStepsOnMaterialized executes query steps on materialized data
func (e *Engine) executeStepsOnMaterialized(ctx *MaterializedContext, steps []parser.Step, current interface{}) ([]Match, error) {
	var matches []Match

	if len(steps) == 0 {
		match := Match{
			Value: current,
			Path:  "",
		}
		if rawBytes, err := json.Marshal(current); err == nil {
			match.Raw = rawBytes
		}
		matches = append(matches, match)
		return matches, nil
	}

	step := steps[0]
	remaining := steps[1:]

	switch step.Type {
	case parser.StepChild:
		return e.executeChildStepOnMaterialized(ctx, step, remaining, current)
	case parser.StepDescendant:
		return e.executeDescendantStepOnMaterialized(ctx, step, remaining, current)
	case parser.StepWildcard:
		return e.executeWildcardStepOnMaterialized(ctx, step, remaining, current)
	default:
		return matches, fmt.Errorf("unsupported step type: %v", step.Type)
	}
}

// executeChildStepOnMaterialized executes a child step on materialized data
func (e *Engine) executeChildStepOnMaterialized(ctx *MaterializedContext, step parser.Step, remaining []parser.Step, current interface{}) ([]Match, error) {
	var matches []Match

	switch v := current.(type) {
	case map[string]interface{}:
		if step.Name != "" {
			if value, exists := v[step.Name]; exists {
				if len(remaining) == 0 {
					match := Match{
						Value: value,
						Path:  step.Name,
					}
					if rawBytes, err := json.Marshal(value); err == nil {
						match.Raw = rawBytes
					}
					matches = append(matches, match)
				} else {
					subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, value)
					if err != nil {
						return nil, err
					}
					for _, match := range subMatches {
						match.Path = step.Name + "." + match.Path
						matches = append(matches, match)
					}
				}
			}
		}
	case []interface{}:
		// Handle array access with predicates
		return e.executeArrayStepOnMaterialized(ctx, step, remaining, v)
	}

	return matches, nil
}

// executeDescendantStepOnMaterialized executes a descendant step on materialized data
func (e *Engine) executeDescendantStepOnMaterialized(ctx *MaterializedContext, step parser.Step, remaining []parser.Step, current interface{}) ([]Match, error) {
	var matches []Match

	// Recursively search through the structure
	var search func(interface{}, string) error
	search = func(value interface{}, path string) error {
		switch v := value.(type) {
		case map[string]interface{}:
			if step.Name != "" {
				if target, exists := v[step.Name]; exists {
					currentPath := path
					if currentPath != "" {
						currentPath += "." + step.Name
					} else {
						currentPath = step.Name
					}

					if len(remaining) == 0 {
						match := Match{
							Value: target,
							Path:  currentPath,
						}
						if rawBytes, err := json.Marshal(target); err == nil {
							match.Raw = rawBytes
						}
						matches = append(matches, match)
					} else {
						subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, target)
						if err != nil {
							return err
						}
						for _, match := range subMatches {
							match.Path = currentPath + "." + match.Path
							matches = append(matches, match)
						}
					}
				}
			}

			// Continue searching in nested objects
			for key, val := range v {
				newPath := path
				if newPath != "" {
					newPath += "." + key
				} else {
					newPath = key
				}
				if err := search(val, newPath); err != nil {
					return err
				}
			}
		case []interface{}:
			// Search in array elements
			for i, val := range v {
				newPath := path + "[" + strconv.Itoa(i) + "]"
				if err := search(val, newPath); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := search(current, ""); err != nil {
		return nil, err
	}

	return matches, nil
}

// executeWildcardStepOnMaterialized executes a wildcard step on materialized data
func (e *Engine) executeWildcardStepOnMaterialized(ctx *MaterializedContext, step parser.Step, remaining []parser.Step, current interface{}) ([]Match, error) {
	var matches []Match

	switch v := current.(type) {
	case map[string]interface{}:
		// Wildcard matches all object keys
		for key, value := range v {
			if len(remaining) == 0 {
				match := Match{
					Value: value,
					Path:  key,
				}
				if rawBytes, err := json.Marshal(value); err == nil {
					match.Raw = rawBytes
				}
				matches = append(matches, match)
			} else {
				subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, value)
				if err != nil {
					return nil, err
				}
				for _, match := range subMatches {
					match.Path = key + "." + match.Path
					matches = append(matches, match)
				}
			}
		}
	case []interface{}:
		// Wildcard matches all array elements
		for i, value := range v {
			if len(remaining) == 0 {
				match := Match{
					Value: value,
					Path:  fmt.Sprintf("[%d]", i),
				}
				if rawBytes, err := json.Marshal(value); err == nil {
					match.Raw = rawBytes
				}
				matches = append(matches, match)
			} else {
				subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, value)
				if err != nil {
					return nil, err
				}
				for _, match := range subMatches {
					match.Path = fmt.Sprintf("[%d]", i) + "." + match.Path
					matches = append(matches, match)
				}
			}
		}
	}

	return matches, nil
}

// executeArrayStepOnMaterialized executes array access on materialized data
func (e *Engine) executeArrayStepOnMaterialized(ctx *MaterializedContext, step parser.Step, remaining []parser.Step, array []interface{}) ([]Match, error) {
	var matches []Match

	for _, predicate := range step.Predicates {
		switch predicate.Type {
		case parser.PredicateIndex:
			// Simple array index access
			index := predicate.Index
			if index < 0 {
				index = len(array) + index // Handle negative indices
			}

			if index >= 0 && index < len(array) {
				value := array[index]
				if len(remaining) == 0 {
					match := Match{
						Value: value,
						Path:  fmt.Sprintf("[%d]", predicate.Index),
					}
					if rawBytes, err := json.Marshal(value); err == nil {
						match.Raw = rawBytes
					}
					matches = append(matches, match)
				} else {
					subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, value)
					if err != nil {
						return nil, err
					}
					for _, match := range subMatches {
						match.Path = fmt.Sprintf("[%d]", predicate.Index) + "." + match.Path
						matches = append(matches, match)
					}
				}
			}
		case parser.PredicateSlice:
			// Array slice access
			start := predicate.Start
			end := predicate.End

			if start < 0 {
				start = len(array) + start
			}
			if end < 0 {
				end = len(array) + end
			}
			if end > len(array) {
				end = len(array)
			}

			if start >= 0 && start < len(array) && end > start {
				for i := start; i < end; i++ {
					value := array[i]
					if len(remaining) == 0 {
						match := Match{
							Value: value,
							Path:  fmt.Sprintf("[%d]", i),
						}
						if rawBytes, err := json.Marshal(value); err == nil {
							match.Raw = rawBytes
						}
						matches = append(matches, match)
					} else {
						subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, value)
						if err != nil {
							return nil, err
						}
						for _, match := range subMatches {
							match.Path = fmt.Sprintf("[%d]", i) + "." + match.Path
							matches = append(matches, match)
						}
					}
				}
			}
		case parser.PredicateWildcard:
			// All array elements
			for i, value := range array {
				if len(remaining) == 0 {
					match := Match{
						Value: value,
						Path:  fmt.Sprintf("[%d]", i),
					}
					if rawBytes, err := json.Marshal(value); err == nil {
						match.Raw = rawBytes
					}
					matches = append(matches, match)
				} else {
					subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, value)
					if err != nil {
						return nil, err
					}
					for _, match := range subMatches {
						match.Path = fmt.Sprintf("[%d]", i) + "." + match.Path
						matches = append(matches, match)
					}
				}
			}
		case parser.PredicateExpression:
			// Filter expression like [?(@.price < 20)]
			rootData := ctx.data

			filteredMap, err := ApplyFilter(array, predicate, rootData)
			if err != nil {
				return nil, fmt.Errorf("error applying filter: %w", err)
			}
			for originalIndex, item := range filteredMap {
				if len(remaining) == 0 {
					match := Match{
						Value: item,
						Path:  fmt.Sprintf("[%d]", originalIndex),
					}
					if rawBytes, err := json.Marshal(item); err == nil {
						match.Raw = rawBytes
					}
					matches = append(matches, match)
				} else {
					subMatches, err := e.executeStepsOnMaterialized(ctx, remaining, item)
					if err != nil {
						return nil, err
					}
					for _, match := range subMatches {
						match.Path = fmt.Sprintf("[%d]", originalIndex) + "." + match.Path
						matches = append(matches, match)
					}
				}
			}
		}
	}

	return matches, nil
}

// ParseSimplePath parses a simple dot-notation path like "store.book.title"
// This is a temporary helper for basic functionality
func ParseSimplePath(path string) []string {
	if path == "" {
		return []string{}
	}

	// Handle array indices in the path
	parts := strings.Split(path, ".")
	var result []string

	for _, part := range parts {
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Split on array access
			openBracket := strings.Index(part, "[")
			if openBracket > 0 {
				result = append(result, part[:openBracket])
			}
			result = append(result, part[openBracket:])
		} else {
			result = append(result, part)
		}
	}

	return result
}

// parseJSONValue parses a JSON value from bytes
func parseJSONValue(data []byte) (interface{}, error) {
	var value interface{}
	err := json.Unmarshal(data, &value)
	return value, err
}

// GetValueBySimplePath is a helper function for basic path access
// This provides immediate functionality while the full XPath parser is being developed
func GetValueBySimplePath(data interface{}, path string) (interface{}, bool) {
	if path == "" {
		return data, true
	}

	// Handle recursive queries (..)
	if strings.Contains(path, "..") {
		return getValueByRecursivePath(data, path)
	}

	parts := ParseSimplePath(path)
	current := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle array access (including slices and negative indices)
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			current = handleArrayAccess(current, part)
			if current == nil {
				return nil, false
			}
		} else {
			// Handle object access
			switch v := current.(type) {
			case map[string]interface{}:
				if value, exists := v[part]; exists {
					current = value
				} else {
					return nil, false
				}
			default:
				return nil, false
			}
		}
	}

	return current, true
}

// ConvertValue converts a value to the specified type
func ConvertValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// Handle JSON number conversion
	switch targetType.Kind() {
	case reflect.String:
		if s, ok := value.(string); ok {
			return s, nil
		}
		return fmt.Sprintf("%v", value), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case float64:
			return int64(v), nil
		case int:
			return int64(v), nil
		case int64:
			return v, nil
		case string:
			return strconv.ParseInt(v, 10, 64)
		}

	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			return strconv.ParseFloat(v, 64)
		}

	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(v)
		}
	}

	return value, nil
}

// GetValueBySimplePathFromRaw extracts a value from raw JSON bytes without materialization
// This is a basic implementation for lazy parsing - more efficient implementation needed
func GetValueBySimplePathFromRaw(data []byte, path string) (interface{}, bool) {
	if path == "" {
		// Return the entire JSON as interface{}
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, false
		}
		return result, true
	}

	// For now, we'll use a simple approach that still unmarshals
	// but only the portion we need. In a full implementation,
	// this would use true streaming/lazy parsing
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}

	return GetValueBySimplePath(result, path)
}

// handleArrayAccess processes array access expressions including slices and negative indices
func handleArrayAccess(current interface{}, arrayExpr string) interface{} {
	if !strings.HasPrefix(arrayExpr, "[") || !strings.HasSuffix(arrayExpr, "]") {
		return nil
	}

	inner := arrayExpr[1 : len(arrayExpr)-1]

	switch v := current.(type) {
	case []interface{}:
		// Handle slice notation [start:end]
		if strings.Contains(inner, ":") {
			return handleArraySlice(v, inner)
		}

		// Handle single index (including negative)
		index, err := strconv.Atoi(inner)
		if err != nil {
			return nil
		}

		if index < 0 {
			index = len(v) + index
		}

		if index >= 0 && index < len(v) {
			return v[index]
		}
		return nil

	default:
		return nil
	}
}

// handleArraySlice processes array slice expressions like [1:3], [:2], [1:]
func handleArraySlice(arr []interface{}, sliceExpr string) interface{} {
	parts := strings.Split(sliceExpr, ":")
	if len(parts) != 2 {
		return nil
	}

	var start, end int
	var err error

	// Parse start index
	if parts[0] == "" {
		start = 0
	} else {
		start, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil
		}
		if start < 0 {
			start = len(arr) + start
		}
	}

	// Parse end index
	if parts[1] == "" {
		end = len(arr)
	} else {
		end, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil
		}
		if end < 0 {
			end = len(arr) + end
		}
	}

	// Bounds checking
	if start < 0 {
		start = 0
	}
	if end > len(arr) {
		end = len(arr)
	}
	if start >= end {
		return []interface{}{}
	}

	return arr[start:end]
}

// getValueByRecursivePath handles recursive queries with ".." syntax
func getValueByRecursivePath(data interface{}, path string) (interface{}, bool) {
	// Split on the first ".." occurrence
	parts := strings.SplitN(path, "..", 2)
	if len(parts) != 2 {
		return nil, false
	}

	prefix := parts[0]
	suffix := parts[1]

	// Remove leading dots from suffix
	suffix = strings.TrimPrefix(suffix, ".")

	var results []interface{}

	// If there's a prefix, navigate to it first
	var startData interface{} = data
	if prefix != "" && prefix != "." {
		if value, exists := GetValueBySimplePath(data, prefix); exists {
			startData = value
		} else {
			return nil, false
		}
	}

	// Recursively search for the suffix pattern
	collectRecursiveMatches(startData, suffix, &results)

	if len(results) == 0 {
		return nil, false
	}

	// If only one result, return it directly; otherwise return the array
	if len(results) == 1 {
		return results[0], true
	}
	return results, true
}

// collectRecursiveMatches recursively collects all matches for the given path
func collectRecursiveMatches(data interface{}, targetPath string, results *[]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		// First check if current level has the target
		if targetPath != "" {
			if value, exists := GetValueBySimplePath(v, targetPath); exists {
				*results = append(*results, value)
			}
		}

		// Recursively search in all values
		for _, value := range v {
			collectRecursiveMatches(value, targetPath, results)
		}

	case []interface{}:
		// Recursively search in all array elements
		for _, item := range v {
			collectRecursiveMatches(item, targetPath, results)
		}
	}
}

// ApplyFilter applies a filter expression to an array of items and returns a map of original indices to matched items
func ApplyFilter(items []interface{}, predicate parser.Predicate, rootData interface{}) (map[int]interface{}, error) {
	if predicate.Type != parser.PredicateExpression {
		results := make(map[int]interface{}, len(items))
		for i, item := range items {
			results[i] = item
		}
		return results, nil
	}

	results := make(map[int]interface{})
	for i, item := range items {
		match, err := evaluateFilterExpression(predicate.Expression, item, rootData)
		if err != nil {
			return nil, err
		}
		if match {
			results[i] = item
		}
	}
	return results, nil
}

// evaluateFilterExpression evaluates a filter expression against an item
func evaluateFilterExpression(expr parser.Expression, currentItem, rootData interface{}) (bool, error) {
	switch expr.Type {
	case parser.ExpressionBinary:
		return evaluateBinaryExpression(expr, currentItem, rootData)
	case parser.ExpressionLiteral:
		return evaluateLiteralExpression(expr)
	case parser.ExpressionPath:
		return evaluatePathExpression(expr, currentItem, rootData)
	default:
		return false, fmt.Errorf("unsupported expression type in filter: %d", expr.Type)
	}
}

// evaluateBinaryExpression evaluates binary operations in filters
func evaluateBinaryExpression(expr parser.Expression, currentItem, rootData interface{}) (bool, error) {
	switch expr.Operator {
	case "&&":
		left, err := evaluateFilterExpression(*expr.Left, currentItem, rootData)
		if err != nil || !left {
			return false, err
		}
		return evaluateFilterExpression(*expr.Right, currentItem, rootData)

	case "||":
		left, err := evaluateFilterExpression(*expr.Left, currentItem, rootData)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil
		}
		return evaluateFilterExpression(*expr.Right, currentItem, rootData)

	default:
		// Comparison operators
		leftValue, err := getFilterExpressionValue(*expr.Left, currentItem, rootData)
		if err != nil {
			return false, err
		}
		rightValue, err := getFilterExpressionValue(*expr.Right, currentItem, rootData)
		if err != nil {
			return false, err
		}

		return compareFilterValues(leftValue, rightValue, expr.Operator)
	}
} // evaluateLiteralExpression evaluates literal values in filters
func evaluateLiteralExpression(expr parser.Expression) (bool, error) {
	switch v := expr.Value.(type) {
	case bool:
		return v, nil
	case float64:
		return v != 0, nil
	case string:
		return v != "", nil
	default:
		return false, fmt.Errorf("unsupported literal type: %T", v)
	}
}

// evaluatePathExpression evaluates path expressions in filters
func evaluatePathExpression(expr parser.Expression, currentItem, rootData interface{}) (bool, error) {
	value, err := getFilterExpressionValue(expr, currentItem, rootData)
	if err != nil {
		return false, err
	}

	// Check if value is truthy
	switch v := value.(type) {
	case bool:
		return v, nil
	case float64:
		return v != 0, nil
	case string:
		return v != "", nil
	case nil:
		return false, nil
	default:
		return true, nil
	}
}

// getFilterExpressionValue gets the actual value from a filter expression
func getFilterExpressionValue(expr parser.Expression, currentItem, rootData interface{}) (interface{}, error) {
	switch expr.Type {
	case parser.ExpressionLiteral:
		return expr.Value, nil

	case parser.ExpressionPath:
		pathStr := strings.Join(expr.Path, ".")

		// Handle @ (current) and $ (root) contexts
		if len(expr.Path) > 0 {
			if expr.Path[0] == "@" {
				if len(expr.Path) == 1 {
					return currentItem, nil
				}
				subPath := strings.Join(expr.Path[1:], ".")
				value, _ := GetValueBySimplePath(currentItem, subPath)
				return value, nil
			} else if expr.Path[0] == "$" {
				if len(expr.Path) == 1 {
					return rootData, nil
				}
				subPath := strings.Join(expr.Path[1:], ".")
				value, _ := GetValueBySimplePath(rootData, subPath)
				return value, nil
			}
		}

		// Default to current item context
		value, _ := GetValueBySimplePath(currentItem, pathStr)
		return value, nil

	default:
		return nil, fmt.Errorf("cannot get value from expression type: %d", expr.Type)
	}
} // compareFilterValues compares two values in a filter context
func compareFilterValues(left, right interface{}, operator string) (bool, error) {
	// Handle nil comparisons
	if left == nil || right == nil {
		switch operator {
		case "==":
			return left == right, nil
		case "!=":
			return left != right, nil
		default:
			return false, nil
		}
	}

	// Handle boolean comparisons with truthiness
	if isBooleanValue(left) || isBooleanValue(right) {
		leftBool := toBooleanValue(left)
		rightBool := toBooleanValue(right)
		switch operator {
		case "==":
			return leftBool == rightBool, nil
		case "!=":
			return leftBool != rightBool, nil
		default:
			return false, fmt.Errorf("unsupported boolean operator: %s", operator)
		}
	}

	// Try to normalize types for comparison
	leftFloat, leftOk := convertToFloat(left)
	rightFloat, rightOk := convertToFloat(right)

	if leftOk && rightOk {
		// Numeric comparison
		switch operator {
		case "==":
			return leftFloat == rightFloat, nil
		case "!=":
			return leftFloat != rightFloat, nil
		case "<":
			return leftFloat < rightFloat, nil
		case "<=":
			return leftFloat <= rightFloat, nil
		case ">":
			return leftFloat > rightFloat, nil
		case ">=":
			return leftFloat >= rightFloat, nil
		}
	}

	// String comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	switch operator {
	case "==":
		return leftStr == rightStr, nil
	case "!=":
		return leftStr != rightStr, nil
	case "<":
		return leftStr < rightStr, nil
	case "<=":
		return leftStr <= rightStr, nil
	case ">":
		return leftStr > rightStr, nil
	case ">=":
		return leftStr >= rightStr, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// convertToFloat attempts to convert a value to float64
func convertToFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// isBooleanValue checks if a value should be treated as boolean
func isBooleanValue(v interface{}) bool {
	_, ok := v.(bool)
	return ok
}

// toBooleanValue converts a value to boolean using truthiness rules
func toBooleanValue(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int32:
		return val != 0
	case int64:
		return val != 0
	case float32:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	case nil:
		return false
	default:
		return true // Non-nil values are truthy
	}
}
