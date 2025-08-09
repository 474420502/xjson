package xjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArraySliceOperations(t *testing.T) {
	data := `{
		"store": {
			"books": [
				{"title": "Book 1", "price": 10},
				{"title": "Book 2", "price": 20},
				{"title": "Book 3", "price": 30},
				{"title": "Book 4", "price": 40},
				{"title": "Book 5", "price": 50}
			]
		}
	}`

	root, err := Parse(data)
	assert.NoError(t, err)

	t.Run("Basic slice [1:3]", func(t *testing.T) {
		result := root.Query("/store/books[1:3]/title")
		assert.NoError(t, result.Error())
		titles := result.Strings()
		assert.Equal(t, []string{"Book 2", "Book 3"}, titles)
	})

	t.Run("Slice from start [:3]", func(t *testing.T) {
		result := root.Query("/store/books[:3]/title")
		assert.NoError(t, result.Error())
		titles := result.Strings()
		assert.Equal(t, []string{"Book 1", "Book 2", "Book 3"}, titles)
	})

	t.Run("Slice to end [2:]", func(t *testing.T) {
		result := root.Query("/store/books[2:]/title")
		assert.NoError(t, result.Error())
		titles := result.Strings()
		assert.Equal(t, []string{"Book 3", "Book 4", "Book 5"}, titles)
	})

	t.Run("Negative indices [-2:]", func(t *testing.T) {
		result := root.Query("/store/books[-2:]/title")
		assert.NoError(t, result.Error())
		titles := result.Strings()
		assert.Equal(t, []string{"Book 4", "Book 5"}, titles)
	})

	t.Run("Negative indices [:-2]", func(t *testing.T) {
		result := root.Query("/store/books[:-2]/title")
		assert.NoError(t, result.Error())
		titles := result.Strings()
		assert.Equal(t, []string{"Book 1", "Book 2", "Book 3"}, titles)
	})

	t.Run("Single element slice [2:3]", func(t *testing.T) {
		result := root.Query("/store/books[2:3]/title")
		assert.NoError(t, result.Error())
		titles := result.Strings()
		assert.Equal(t, []string{"Book 3"}, titles)
	})

	t.Run("Empty slice [3:3]", func(t *testing.T) {
		result := root.Query("/store/books[3:3]/title")
		assert.NoError(t, result.Error())
		assert.Equal(t, 0, result.Len())
	})

	t.Run("Invalid slice bounds", func(t *testing.T) {
		result := root.Query("/store/books[5:6]/title")
		assert.Error(t, result.Error())
	})
}
