// Package filter provides expression evaluation for XJSON filter queries.
// It implements a runtime evaluator for parsed filter expressions like [?(@.price < 20)]
package filter

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/474420502/xjson/internal/engine"
	"github.com/474420502/xjson/internal/parser"
)

// EvaluationContext provides context for filter expression evaluation
type EvaluationContext struct {
	CurrentItem interface{} // The @ context (current array item being evaluated)
	RootData    interface{} // The $ context (root document)
}

// FilterEvaluator evaluates filter expressions on JSON data
type FilterEvaluator struct{}

// NewFilterEvaluator creates a new filter evaluator
func NewFilterEvaluator() *FilterEvaluator {
	return &FilterEvaluator{}
}

// EvaluateExpression evaluates a parsed expression against the given context
func (fe *FilterEvaluator) EvaluateExpression(expr parser.Expression, ctx *EvaluationContext) (bool, error) {
	switch expr.Type {
	case parser.ExpressionBinary:
		return fe.evaluateBinaryExpression(expr, ctx)
	case parser.ExpressionUnary:
		return fe.evaluateUnaryExpression(expr, ctx)
	case parser.ExpressionLiteral:
		return fe.evaluateLiteralExpression(expr, ctx)
	case parser.ExpressionPath:
		return fe.evaluatePathExpression(expr, ctx)
	case parser.ExpressionFunction:
		return fe.evaluateFunctionExpression(expr, ctx)
	default:
		return false, fmt.Errorf("unsupported expression type: %d", expr.Type)
	}
}

// evaluateBinaryExpression evaluates binary operations (==, !=, <, >, &&, ||)
func (fe *FilterEvaluator) evaluateBinaryExpression(expr parser.Expression, ctx *EvaluationContext) (bool, error) {
	switch expr.Operator {
	case "&&":
		// Logical AND - short circuit evaluation
		leftResult, err := fe.EvaluateExpression(*expr.Left, ctx)
		if err != nil || !leftResult {
			return false, err
		}
		return fe.EvaluateExpression(*expr.Right, ctx)

	case "||":
		// Logical OR - short circuit evaluation
		leftResult, err := fe.EvaluateExpression(*expr.Left, ctx)
		if err != nil {
			return false, err
		}
		if leftResult {
			return true, nil
		}
		return fe.EvaluateExpression(*expr.Right, ctx)

	default:
		// Comparison operators - evaluate both sides first
		leftValue, err := fe.getExpressionValue(*expr.Left, ctx)
		if err != nil {
			return false, err
		}

		rightValue, err := fe.getExpressionValue(*expr.Right, ctx)
		if err != nil {
			return false, err
		}

		return fe.compareValues(leftValue, rightValue, expr.Operator)
	}
}

// evaluateUnaryExpression evaluates unary operations (!)
func (fe *FilterEvaluator) evaluateUnaryExpression(expr parser.Expression, ctx *EvaluationContext) (bool, error) {
	if expr.Operator == "!" {
		result, err := fe.EvaluateExpression(*expr.Left, ctx)
		return !result, err
	}
	return false, fmt.Errorf("unsupported unary operator: %s", expr.Operator)
}

// evaluateLiteralExpression evaluates literal values (true, false, numbers, strings)
func (fe *FilterEvaluator) evaluateLiteralExpression(expr parser.Expression, ctx *EvaluationContext) (bool, error) {
	// Literals in boolean context: numbers (0 = false, non-0 = true), booleans as-is
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

// evaluatePathExpression evaluates path expressions (@.price, $.root.field)
func (fe *FilterEvaluator) evaluatePathExpression(expr parser.Expression, ctx *EvaluationContext) (bool, error) {
	value, err := fe.getExpressionValue(expr, ctx)
	if err != nil {
		return false, err
	}

	// Path exists and has truthy value
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
		return true, nil // Non-nil values are truthy
	}
}

// evaluateFunctionExpression evaluates function calls (exists, includes, etc.)
func (fe *FilterEvaluator) evaluateFunctionExpression(expr parser.Expression, ctx *EvaluationContext) (bool, error) {
	switch expr.Function {
	case "exists":
		// Check if path exists
		if len(expr.Path) == 0 {
			return false, errors.New("exists() requires a path argument")
		}
		pathStr := strings.Join(expr.Path, ".")
		_, exists := engine.GetValueBySimplePath(ctx.CurrentItem, pathStr)
		return exists, nil

	case "includes":
		// Check if array includes value
		if expr.Left == nil || expr.Right == nil {
			return false, errors.New("includes() requires two arguments")
		}

		// Get the array value
		arrayValue, err := fe.getExpressionValue(*expr.Left, ctx)
		if err != nil {
			return false, err
		}

		// Get the search value
		searchValue, err := fe.getExpressionValue(*expr.Right, ctx)
		if err != nil {
			return false, err
		}

		// Check if array contains the search value
		switch arr := arrayValue.(type) {
		case []interface{}:
			for _, item := range arr {
				if fe.valuesEqual(item, searchValue) {
					return true, nil
				}
			}
			return false, nil
		case string:
			// String contains check
			searchStr, ok := searchValue.(string)
			if !ok {
				return false, nil
			}
			return strings.Contains(arr, searchStr), nil
		default:
			return false, nil
		}

	default:
		return false, fmt.Errorf("unsupported function: %s", expr.Function)
	}
}

// valuesEqual compares two values for equality
func (fe *FilterEvaluator) valuesEqual(a, b interface{}) bool {
	// 宽松等值比较：数值、布尔、nil、字符串等
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// 数值比较
	if fe.isNumber(a) && fe.isNumber(b) {
		af, _ := fe.toFloat64(a)
		bf, _ := fe.toFloat64(b)
		return af == bf
	}
	// 布尔与数值/布尔
	if fe.isBool(a) || fe.isBool(b) {
		return fe.toBool(a) == fe.toBool(b)
	}
	// 字符串
	if fe.isString(a) && fe.isString(b) {
		return a == b
	}
	// 其它类型 fallback
	return reflect.DeepEqual(a, b)
}

// getExpressionValue gets the actual value of an expression (not boolean result)
func (fe *FilterEvaluator) getExpressionValue(expr parser.Expression, ctx *EvaluationContext) (interface{}, error) {
	switch expr.Type {
	case parser.ExpressionLiteral:
		return expr.Value, nil

	case parser.ExpressionPath:
		pathStr := strings.Join(expr.Path, ".")

		// Handle @ (current item) and $ (root) contexts
		if len(expr.Path) > 0 && expr.Path[0] == "@" {
			// Remove @ prefix and evaluate on current item
			if len(expr.Path) == 1 {
				return ctx.CurrentItem, nil
			}
			subPath := strings.Join(expr.Path[1:], ".")
			value, _ := engine.GetValueBySimplePath(ctx.CurrentItem, subPath)
			return value, nil
		} else if len(expr.Path) > 0 && expr.Path[0] == "$" {
			// Remove $ prefix and evaluate on root
			if len(expr.Path) == 1 {
				return ctx.RootData, nil
			}
			subPath := strings.Join(expr.Path[1:], ".")
			value, _ := engine.GetValueBySimplePath(ctx.RootData, subPath)
			return value, nil
		} else {
			// Evaluate on current item by default
			value, _ := engine.GetValueBySimplePath(ctx.CurrentItem, pathStr)
			return value, nil
		}

	default:
		return nil, fmt.Errorf("cannot get value from expression type: %d", expr.Type)
	}
}

// compareValues compares two values using the given operator
func (fe *FilterEvaluator) compareValues(left, right interface{}, operator string) (bool, error) {
	// Handle nil values
	if left == nil || right == nil {
		switch operator {
		case "==":
			return left == right, nil
		case "!=":
			return left != right, nil
		default:
			return false, nil // nil values can't be ordered
		}
	}

	// Convert values to comparable types
	leftVal, rightVal, err := fe.normalizeForComparison(left, right)
	if err != nil {
		return false, err
	}

	switch operator {
	case "==":
		return fe.equals(leftVal, rightVal), nil
	case "!=":
		return !fe.equals(leftVal, rightVal), nil
	case "<":
		return fe.lessThan(leftVal, rightVal)
	case "<=":
		return fe.lessThanOrEqual(leftVal, rightVal)
	case ">":
		return fe.greaterThan(leftVal, rightVal)
	case ">=":
		return fe.greaterThanOrEqual(leftVal, rightVal)
	default:
		return false, fmt.Errorf("unsupported comparison operator: %s", operator)
	}
}

// normalizeForComparison converts values to comparable types
func (fe *FilterEvaluator) normalizeForComparison(left, right interface{}) (interface{}, interface{}, error) {
	// If both are numbers, convert to float64
	if fe.isNumber(left) && fe.isNumber(right) {
		leftFloat, _ := fe.toFloat64(left)
		rightFloat, _ := fe.toFloat64(right)
		return leftFloat, rightFloat, nil
	}

	// If both are strings, keep as strings
	if fe.isString(left) && fe.isString(right) {
		return left, right, nil
	}

	// If both are booleans, keep as booleans
	if fe.isBool(left) && fe.isBool(right) {
		return left, right, nil
	}

	// Handle boolean/number comparison (JavaScript-like truthiness)
	if fe.isBool(left) || fe.isBool(right) {
		leftBool := fe.toBool(left)
		rightBool := fe.toBool(right)
		return leftBool, rightBool, nil
	}

	// Try to convert both to strings for comparison
	return fmt.Sprintf("%v", left), fmt.Sprintf("%v", right), nil
}

// Helper methods for type checking and conversion
func (fe *FilterEvaluator) isNumber(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func (fe *FilterEvaluator) isString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

func (fe *FilterEvaluator) isBool(v interface{}) bool {
	_, ok := v.(bool)
	return ok
}

func (fe *FilterEvaluator) toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int8:
		return val != 0
	case int16:
		return val != 0
	case int32:
		return val != 0
	case int64:
		return val != 0
	case uint:
		return val != 0
	case uint8:
		return val != 0
	case uint16:
		return val != 0
	case uint32:
		return val != 0
	case uint64:
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

func (fe *FilterEvaluator) toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func (fe *FilterEvaluator) equals(left, right interface{}) bool {
	return reflect.DeepEqual(left, right)
}

func (fe *FilterEvaluator) lessThan(left, right interface{}) (bool, error) {
	if fe.isNumber(left) && fe.isNumber(right) {
		leftFloat, _ := fe.toFloat64(left)
		rightFloat, _ := fe.toFloat64(right)
		return leftFloat < rightFloat, nil
	}
	if fe.isString(left) && fe.isString(right) {
		return left.(string) < right.(string), nil
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

func (fe *FilterEvaluator) lessThanOrEqual(left, right interface{}) (bool, error) {
	eq := fe.equals(left, right)
	if eq {
		return true, nil
	}
	return fe.lessThan(left, right)
}

func (fe *FilterEvaluator) greaterThan(left, right interface{}) (bool, error) {
	result, err := fe.lessThanOrEqual(left, right)
	return !result, err
}

func (fe *FilterEvaluator) greaterThanOrEqual(left, right interface{}) (bool, error) {
	result, err := fe.lessThan(left, right)
	return !result, err
}
