package xjson

import (
	"testing"
)

func BenchmarkSimpleQuery(b *testing.B) {
	jsonData := `{"user": {"name": "John", "age": 30}}`
	doc, _ := ParseString(jsonData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Query("user.name")
	}
}

func BenchmarkFilterQuery(b *testing.B) {
	jsonData := `{
		"products": [
			{"name": "Laptop", "price": 999.99},
			{"name": "Mouse", "price": 25.50},
			{"name": "Chair", "price": 85.00}
		]
	}`
	doc, _ := ParseString(jsonData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Query("products[?(@.price < 100)]")
	}
}

func BenchmarkRecursiveQuery(b *testing.B) {
	jsonData := `{
		"store": {
			"products": [
				{"price": 10.0},
				{"price": 20.0}
			],
			"info": {"price": 5.0}
		}
	}`
	doc, _ := ParseString(jsonData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Query("..price")
	}
}

func BenchmarkArraySlice(b *testing.B) {
	jsonData := `{
		"items": [
			{"id": 1}, {"id": 2}, {"id": 3}, {"id": 4}, {"id": 5}
		]
	}`
	doc, _ := ParseString(jsonData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.Query("items[1:3]")
	}
}

func BenchmarkWriteOperation(b *testing.B) {
	jsonData := `{"user": {"name": "John", "age": 30}}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := ParseString(jsonData)
		_ = doc.Set("user.age", 31)
	}
}
