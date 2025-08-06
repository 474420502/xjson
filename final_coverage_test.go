package xjson

import (
	"testing"
)

func TestFinalCoverageImprovements(t *testing.T) {
	// Test String() method edge cases
	t.Run("String_EdgeCases", func(t *testing.T) {
		// Test complex type JSON marshaling
		doc, _ := ParseString(`{"complex":{"nested":{"array":[1,2,3]}}}`)
		complex := doc.Query("/complex")
		str, err := complex.String()
		if err != nil {
			t.Errorf("String() on complex type should succeed, got error: %v", err)
		}
		if str == "" {
			t.Error("String() on complex type should return JSON representation")
		}
	})

	// Test Float() method edge cases
	t.Run("Float_EdgeCases", func(t *testing.T) {
		// Test bool type conversion - should fail
		doc, _ := ParseString(`{"bool":true}`)
		_, err := doc.Query("/bool").Float()
		if err == nil {
			t.Error("Float() on bool should return error")
		}

		// Test array type conversion - should fail
		doc2, _ := ParseString(`{"arr":[1,2,3]}`)
		_, err2 := doc2.Query("/arr").Float()
		if err2 == nil {
			t.Error("Float() on array should return error")
		}

		// Test invalid string to float conversion
		doc3, _ := ParseString(`{"str":"not_a_number"}`)
		_, err3 := doc3.Query("/str").Float()
		if err3 == nil {
			t.Error("Float() on invalid string should return error")
		}
	})

	// Test Index() edge cases
	t.Run("Index_EdgeCases", func(t *testing.T) {
		// Test index on object (should fail)
		doc, _ := ParseString(`{"obj":{"a":1}}`)
		result := doc.Query("/obj").Index(0)
		if result.Exists() {
			t.Error("Index() on object should return non-existent result")
		}

		// Test index out of bounds (negative)
		doc2, _ := ParseString(`{"arr":[1,2,3]}`)
		result2 := doc2.Query("/arr").Index(-10)
		if result2.Exists() {
			t.Error("Index() with very negative index should return non-existent result")
		}
	})

	// Test Int64() edge cases
	t.Run("Int64_EdgeCases", func(t *testing.T) {
		// Test array type conversion - should fail
		doc, _ := ParseString(`{"arr":[1,2,3]}`)
		_, err := doc.Query("/arr").Int64()
		if err == nil {
			t.Error("Int64() on array should return error")
		}

		// Test object type conversion - should fail
		doc2, _ := ParseString(`{"obj":{"a":1}}`)
		_, err2 := doc2.Query("/obj").Int64()
		if err2 == nil {
			t.Error("Int64() on object should return error")
		}

		// Test null value
		doc3, _ := ParseString(`{"null":null}`)
		_, err3 := doc3.Query("/null").Int64()
		if err3 == nil {
			t.Error("Int64() on null should return error")
		}
	})

	// Test Bool() edge cases
	t.Run("Bool_EdgeCases", func(t *testing.T) {
		// Test array type conversion - should now fail
		doc, _ := ParseString(`{"arr":[1,2,3]}`)
		_, err := doc.Query("/arr").Bool()
		if err == nil {
			t.Error("Bool() on array should fail with type mismatch")
		}

		// Test object type conversion - should now fail
		doc2, _ := ParseString(`{"obj":{"a":1}}`)
		_, err2 := doc2.Query("/obj").Bool()
		if err2 == nil {
			t.Error("Bool() on object should fail with type mismatch")
		}

		// Test string that can't be parsed as bool - should now fail
		doc3, _ := ParseString(`{"str":"not_bool_but_non_empty"}`)
		_, err3 := doc3.Query("/str").Bool()
		if err3 == nil {
			t.Error("Bool() on non-boolean string should fail with type mismatch")
		}

		// Test empty string - should now fail
		doc4, _ := ParseString(`{"str":""}`)
		_, err4 := doc4.Query("/str").Bool()
		if err4 == nil {
			t.Error("Bool() on empty string should fail with type mismatch")
		}
	})

	// Test MustBool() edge cases
	t.Run("MustBool_EdgeCases", func(t *testing.T) {
		// Test MustBool on error result (should panic)
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustBool() on non-existent should panic")
			}
		}()
		doc, _ := ParseString(`{"num":123}`)
		doc.Query("non_existent").MustBool()
	})

	// Test Query() edge cases for better coverage
	t.Run("Query_EdgeCases", func(t *testing.T) {
		// Test invalid document query
		doc := &Document{err: ErrInvalidJSON}
		result := doc.Query("/test")
		if result.Exists() {
			t.Error("Query on invalid document should return non-existent result")
		}

		// Test complex path that requires full parser
		doc2, _ := ParseString(`{"store":{"books":[{"title":"Book1"},{"title":"Book2"}]}}`)
		result2 := doc2.Query("/store/books[title == 'Book1']")
		// This should use the complex path parser, not simple path
		if result2.Exists() {
			t.Log("Complex query succeeded (implementation may support filters)")
		} else {
			t.Log("Complex query failed (expected for current implementation)")
		}
	})

	// Test ForEach() edge cases
	t.Run("ForEach_EdgeCases", func(t *testing.T) {
		// Test ForEach with error result
		doc := &Document{err: ErrInvalidJSON}
		result := doc.Query("test")
		called := false
		result.ForEach(func(i int, v IResult) bool {
			called = true
			return true
		})
		if called {
			t.Error("ForEach on error result should not call callback")
		}
	})

	// Test Map() edge cases
	t.Run("Map_EdgeCases", func(t *testing.T) {
		// Test Map with error result
		doc := &Document{err: ErrInvalidJSON}
		result := doc.Query("test")
		mapped := result.Map(func(i int, v IResult) interface{} {
			return "test"
		})
		if mapped != nil {
			t.Error("Map on error result should return nil")
		}
	})

	// Test Filter() edge cases
	t.Run("Filter_EdgeCases", func(t *testing.T) {
		// Test Filter with error result
		doc := &Document{err: ErrInvalidJSON}
		result := doc.Query("test")
		filtered := result.Filter(func(i int, v IResult) bool {
			return true
		})
		if filtered.Exists() {
			t.Error("Filter on error result should return non-existent result")
		}
	})
}
