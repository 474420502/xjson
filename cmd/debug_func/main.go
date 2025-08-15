package main

import (
	"fmt"

	"github.com/474420502/xjson/internal/core"
	"github.com/474420502/xjson/internal/engine"
)

func main() {
	jsonData := []byte(`{
		"books": [
			{"title": "Moby Dick", "price": 8.99},
			{"title": "Clean Code", "price": 29.99},
			{"title": "The Hobbit", "price": 12.99}
		]
	}`)
	root, err := engine.Parse(jsonData)
	if err != nil {
		panic(err)
	}

	root.RegisterFunc("cheap", func(n core.Node) core.Node {
		return n.Filter(func(child core.Node) bool {
			price, ok := child.Get("price").RawFloat()
			return ok && price < 20
		})
	})

	filtered := root.Query("/books[@cheap]")
	fmt.Println("Filtered raw:", filtered.Raw(), "valid:", filtered.IsValid(), "type:", filtered.Type())
	if filtered.IsValid() && filtered.Type() == core.Array {
		arr := filtered.Array()
		for i, v := range arr {
			fmt.Printf("elem %d: type=%v raw=%q\n", i, v.Type(), v.Raw())
		}
	}
	// Now query titles
	res := root.Query("/books[@cheap]/title")
	fmt.Println("Title query result raw:", res.Raw(), "valid:", res.IsValid())
	if res.IsValid() {
		fmt.Println("Strings:", res.Strings())
	} else {
		fmt.Println("Invalid:", res.Error())
	}
}
