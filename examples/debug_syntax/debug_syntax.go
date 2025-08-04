package main

import (
	"fmt"

	xjson "github.com/474420502/xjson"
)

func main() {
	doc, _ := xjson.ParseString(`{"a":[{"b":2}],"a.b.c":1}`)
	result := doc.Query("a[?(")
	fmt.Printf("Query 'a[?(' exists: %v\n", result.Exists())
	fmt.Printf("Query 'a[?(' count: %v\n", result.Count())
}
