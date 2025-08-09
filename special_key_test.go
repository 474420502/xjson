package xjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecialKeyHandling(t *testing.T) {
	data := `{
		"data/user-profile": {
			"name": "John Doe",
			"age": 30,
			"contact.email": "john@example.com",
			"contact[address]": {
				"street": "123 Main St",
				"city": "New York"
			},
			"items": [
				{"id": 1, "name": "item1"},
				{"id": 2, "name": "item2"}
			]
		},
		"api/v1/users": [
			{"id": 1, "name": "user1"},
			{"id": 2, "name": "user2"}
		],
		"nested": {
			"key.with.dots": {
				"value": "dots in key"
			},
			"key-with-dashes": {
				"value": "dashes in key"
			}
		}
	}`

	root, err := Parse(data)
	assert.NoError(t, err)

	t.Run("Single quoted key with slash", func(t *testing.T) {
		result := root.Query("['data/user-profile']/name")
		assert.NoError(t, result.Error())
		assert.Equal(t, "John Doe", result.String())
	})

	t.Run("Double quoted key with slash", func(t *testing.T) {
		result := root.Query("[\"data/user-profile\"]/age")
		assert.NoError(t, result.Error())
		assert.Equal(t, "30", result.String())
	})

	t.Run("Key with dot using quotes", func(t *testing.T) {
		result := root.Query("['data/user-profile']/['contact.email']")
		assert.NoError(t, result.Error())
		assert.Equal(t, "john@example.com", result.String())
	})

	t.Run("Key with brackets using quotes", func(t *testing.T) {
		result := root.Query("['data/user-profile']/['contact[address]']/street")
		assert.NoError(t, result.Error())
		assert.Equal(t, "123 Main St", result.String())
	})

	t.Run("API version path with quotes", func(t *testing.T) {
		result := root.Query("[\"api/v1/users\"][0]/name")
		assert.NoError(t, result.Error())
		assert.Equal(t, "user1", result.String())
	})

	t.Run("Mixed path with quoted and unquoted keys", func(t *testing.T) {
		result := root.Query("['data/user-profile']/items[1]/name")
		assert.NoError(t, result.Error())
		assert.Equal(t, "item2", result.String())
	})

	t.Run("Nested keys with dots", func(t *testing.T) {
		result := root.Query("nested/['key.with.dots']/value")
		assert.NoError(t, result.Error())
		assert.Equal(t, "dots in key", result.String())
	})

	t.Run("Keys with dashes", func(t *testing.T) {
		result := root.Query("nested/['key-with-dashes']/value")
		assert.NoError(t, result.Error())
		assert.Equal(t, "dashes in key", result.String())
	})

	t.Run("Invalid quote combinations", func(t *testing.T) {
		result := root.Query("['data/user-profile\")/name")
		assert.Error(t, result.Error())
	})

	t.Run("Empty quoted key", func(t *testing.T) {
		result := root.Query("['']/name")
		assert.Error(t, result.Error())
	})
}
