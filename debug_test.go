package xjson

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAppend(t *testing.T) {
	jsonData := `{
		"items": [1, 2, 3]
	}`

	root, err := Parse(jsonData)
	assert.NoError(t, err)
	assert.True(t, root.IsValid())

	// Check initial length
	items := root.Get("items")
	assert.Equal(t, 3, items.Len())
	
	// Append a new item
	items.Append(4)
	
	// Check updated length
	items = root.Get("items") // Get fresh reference
	assert.Equal(t, 4, items.Len())
	
	// Check the new item
	assert.Equal(t, int64(4), items.Index(3).Int())
}