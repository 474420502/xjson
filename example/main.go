package main

import (
	"fmt"
	"log"

	"github.com/474420502/xjson"
)

func main() {
	// Test data with nested structure
	data := `{
		"store": {
			"books": [
				{"title": "Moby Dick", "price": 8.99, "tags": ["classic", "adventure"]},
				{"title": "Clean Code", "price": 29.99, "tags": ["programming"]}
			],
			"electronics": {
				"phones": [
					{"title": "iPhone", "price": 999.99}
				]
			}
		},
		"prices": [1.99, 5.99, 10.99]
	}`

	// Parse JSON
	root, err := xjson.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	// Test recursive descent
	fmt.Println("=== Testing Recursive Descent ===")
	titles := root.Query("//title")
	if titles.Error() != nil {
		fmt.Printf("Error: %v\n", titles.Error())
	} else {
		fmt.Printf("Found titles: %v\n", titles.Strings())
	}

	// Test parent navigation
	fmt.Println("\n=== Testing Parent Navigation ===")
	// Navigate to the first book's title
	bookTitle := root.Query("/store/books/0/title")
	if bookTitle.Error() != nil {
		fmt.Printf("Error getting book title: %v\n", bookTitle.Error())
	} else {
		fmt.Printf("Book title: %s\n", bookTitle.String())
		
		// Navigate back to the parent (the book object)
		parent := bookTitle.Query("../")
		if parent.Error() != nil {
			fmt.Printf("Error getting parent: %v\n", parent.Error())
		} else {
			fmt.Printf("Parent type: %v\n", parent.Type())
		}
	}

	// Test combined features
	fmt.Println("\n=== Testing Combined Features ===")
	// Find all prices recursively
	allPrices := root.Query("//price")
	if allPrices.Error() != nil {
		fmt.Printf("Error finding prices: %v\n", allPrices.Error())
	} else {
		fmt.Printf("Found prices: %v\n", allPrices.Array())
	}
}