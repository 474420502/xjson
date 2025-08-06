package modifier

import (
	"errors"
	"fmt"

	"github.com/474420502/xjson/internal/parser"
)

// Modifier handles path-based modifications on materialized JSON data
type Modifier struct{}

// NewModifier creates a new modifier instance
func NewModifier() *Modifier {
	return &Modifier{}
}

// Set sets a value at the specified path in the materialized data
func (m *Modifier) Set(data *interface{}, query *parser.Query, value interface{}) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}

	if len(query.Steps) == 0 {
		*data = value
		return nil
	}

	return m.setAtPath(data, query.Steps, value)
}

// Delete removes a value at the specified path in the materialized data
func (m *Modifier) Delete(data *interface{}, query *parser.Query) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}

	if len(query.Steps) == 0 {
		return errors.New("cannot delete root")
	}

	return m.deleteAtPath(data, query.Steps)
}

func (m *Modifier) setAtPath(data *interface{}, steps []parser.Step, value interface{}) error {
	// If the data is nil, we need to create the root container.
	if *data == nil {
		if len(steps) > 0 && steps[0].Name == "" && len(steps[0].Predicates) > 0 {
			// If the first step is an array access, we can't create it.
			return errors.New("cannot set array element on nil root")
		}
		*data = make(map[string]interface{})
	}

	parent := *data

	// Navigate to the parent element of the target.
	for i := 0; i < len(steps)-1; i++ {
		step := steps[i]

		if step.Name != "" {
			obj, ok := parent.(map[string]interface{})
			if !ok {
				return fmt.Errorf("path error: expected object at step %d, but got %T", i, parent)
			}
			child, exists := obj[step.Name]
			if !exists {
				// Auto-create path
				if i+1 < len(steps) && len(steps[i+1].Predicates) > 0 {
					child = make([]interface{}, 0)
				} else {
					child = make(map[string]interface{})
				}
				obj[step.Name] = child
			}
			parent = child
		}

		if len(step.Predicates) > 0 {
			if len(step.Predicates) > 1 {
				return errors.New("multiple predicates are not supported for set")
			}
			pred := step.Predicates[0]
			if pred.Type != parser.PredicateIndex {
				return errors.New("only index predicates are supported for set")
			}

			arr, ok := parent.([]interface{})
			if !ok {
				return fmt.Errorf("path error: expected array for predicate at step %d, but got %T", i, parent)
			}
			if pred.Index < 0 || pred.Index >= len(arr) {
				return fmt.Errorf("path error: index %d out of bounds at step %d", pred.Index, i)
			}
			parent = arr[pred.Index]
		}
	}

	// Perform the set operation on the final step.
	finalStep := steps[len(steps)-1]

	if finalStep.Name != "" {
		obj, ok := parent.(map[string]interface{})
		if !ok {
			return fmt.Errorf("final step error: expected object, but got %T", parent)
		}

		target := parent
		// If there is a predicate, we need to get/create the array first.
		if len(finalStep.Predicates) > 0 {
			child, exists := obj[finalStep.Name]
			if !exists {
				child = make([]interface{}, 0)
				obj[finalStep.Name] = child
			}
			target = child
		} else {
			obj[finalStep.Name] = value
			return nil
		}

		// If we are here, it means we have predicates on the final step.
		arr, ok := target.([]interface{})
		if !ok {
			return fmt.Errorf("final step error: expected array for predicate, but got %T", target)
		}
		pred := finalStep.Predicates[0]
		if pred.Type != parser.PredicateIndex {
			return errors.New("only index predicates are supported for set")
		}
		idx := pred.Index

		if idx < 0 {
			return fmt.Errorf("final step error: negative index %d not supported for set", idx)
		}

		// Grow the array if necessary.
		if idx >= len(arr) {
			newArr := make([]interface{}, idx+1)
			copy(newArr, arr)
			arr = newArr
			obj[finalStep.Name] = arr // Update the map with the new slice header
		}
		arr[idx] = value

	} else if len(finalStep.Predicates) > 0 {
		arr, ok := parent.([]interface{})
		if !ok {
			return fmt.Errorf("final step error: expected array for predicate, but got %T", parent)
		}
		pred := finalStep.Predicates[0]
		if pred.Type != parser.PredicateIndex {
			return errors.New("only index predicates are supported for set")
		}
		idx := pred.Index

		if idx < 0 || idx >= len(arr) {
			return fmt.Errorf("final step error: index %d out of bounds", idx)
		}
		arr[idx] = value
	}

	return nil
}

func (m *Modifier) deleteAtPath(data *interface{}, steps []parser.Step) error {
	if len(steps) == 0 {
		return errors.New("cannot delete root")
	}

	parent := *data
	for i := 0; i < len(steps)-1; i++ {
		step := steps[i]
		if step.Name != "" {
			obj, ok := parent.(map[string]interface{})
			if !ok {
				return fmt.Errorf("path error: expected object at step %d", i)
			}
			child, exists := obj[step.Name]
			if !exists {
				return fmt.Errorf("path not found at step %d", i)
			}
			parent = child
		}
		if len(step.Predicates) > 0 {
			pred := step.Predicates[0]
			arr, ok := parent.([]interface{})
			if !ok {
				return fmt.Errorf("path error: expected array for predicate at step %d", i)
			}
			if pred.Index < 0 || pred.Index >= len(arr) {
				return fmt.Errorf("path error: index %d out of bounds at step %d", pred.Index, i)
			}
			parent = arr[pred.Index]
		}
	}

	finalStep := steps[len(steps)-1]
	if finalStep.Name != "" {
		obj, ok := parent.(map[string]interface{})
		if !ok {
			return fmt.Errorf("final step error: expected object")
		}
		if len(finalStep.Predicates) == 0 {
			delete(obj, finalStep.Name)
		} else {
			return errors.New("deletion with predicate on final object key not supported")
		}
	} else if len(finalStep.Predicates) > 0 {
		return errors.New("array element deletion not implemented yet")
	}

	return nil
}
