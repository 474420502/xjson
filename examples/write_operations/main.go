package main

import (
	"fmt"
	"log"

	"github.com/474420502/xjson"
)

func main() {
	// Demonstrate Write Operations in XJSON

	fmt.Println("=== XJSON Write Operations Demo ===")
	fmt.Println()

	// Original JSON data
	jsonData := `{
		"user": {
			"id": 123,
			"name": "John Doe",
			"email": "john@example.com",
			"preferences": {
				"theme": "dark",
				"notifications": true
			}
		},
		"posts": [
			{"title": "First Post", "views": 100},
			{"title": "Second Post", "views": 250}
		]
	}`

	// Parse the JSON
	doc, err := xjson.ParseString(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("1. Original JSON:")
	printJSON(doc)

	// Check if materialized (should be false initially)
	fmt.Printf("Is materialized: %t\n\n", doc.IsMaterialized())

	// Read operations don't trigger materialization
	fmt.Println("2. Read operation (Query user.name):")
	name := doc.Query("user.name").MustString()
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Still not materialized: %t\n\n", doc.IsMaterialized())

	// First write operation triggers materialization
	fmt.Println("3. First write operation (Set user.age):")
	err = doc.Set("user.age", 25)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Now materialized: %t\n", doc.IsMaterialized())
	printJSON(doc)

	// Update existing nested property
	fmt.Println("4. Update nested property (user.preferences.theme):")
	err = doc.Set("user.preferences.theme", "light")
	if err != nil {
		log.Fatal(err)
	}
	printJSON(doc)

	// Add new nested property
	fmt.Println("5. Add new nested property (user.preferences.language):")
	err = doc.Set("user.preferences.language", "en")
	if err != nil {
		log.Fatal(err)
	}
	printJSON(doc)

	// Update array element (array access would need to be implemented)
	fmt.Println("6. Update user name:")
	err = doc.Set("user.name", "Jane Smith")
	if err != nil {
		log.Fatal(err)
	}
	printJSON(doc)

	// Delete a property
	fmt.Println("7. Delete user email:")
	err = doc.Delete("user.email")
	if err != nil {
		log.Fatal(err)
	}
	printJSON(doc)

	// Delete nested property
	fmt.Println("8. Delete notification preference:")
	err = doc.Delete("user.preferences.notifications")
	if err != nil {
		log.Fatal(err)
	}
	printJSON(doc)

	// Verify final state with queries
	fmt.Println("9. Final verification with queries:")
	fmt.Printf("User name: %s\n", doc.Query("user.name").MustString())
	fmt.Printf("User age: %d\n", doc.Query("user.age").MustInt())
	fmt.Printf("Theme: %s\n", doc.Query("user.preferences.theme").MustString())
	fmt.Printf("Language: %s\n", doc.Query("user.preferences.language").MustString())
	fmt.Printf("Email exists: %t\n", doc.Query("user.email").Exists())
	fmt.Printf("Notifications exists: %t\n", doc.Query("user.preferences.notifications").Exists())
}

func printJSON(doc *xjson.Document) {
	bytes, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("JSON: %s\n\n", string(bytes))
}
