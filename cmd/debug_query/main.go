package main

import (
	"fmt"

	"github.com/474420502/xjson/internal/engine"
)

func main() {
	jsonData := []byte(`{
		"store": {
			"books": {"title": "Book 1", "price": 10},
			"bikes": {"color": "red", "price": 100}
		}
	}`)
	root, err := engine.Parse(jsonData)
	if err != nil {
		panic(err)
	}
	res := root.Query("/store/*/price")
	fmt.Println("Raw result:", res.Raw())
	if res.IsValid() {
		fmt.Println("Type:", res.Type())
		fmt.Println("Strings:", res.Strings())
	} else {
		fmt.Println("Query invalid:", res.Error())
	}
}
