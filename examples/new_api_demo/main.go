package main

import (
	"fmt"
	"log"

	"github.com/474420502/xjson"
)

func main() {
	// 测试数据
	jsonData := `{
		"store": {
			"book": [
				{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "price": 8.99},
				{"category": "fiction", "author": "J.R.R. Tolkien", "title": "The Lord of the Rings", "price": 22.99},
				{"category": "science", "author": "Carl Sagan", "title": "Cosmos", "price": 15.99}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		}
	}`

	doc, err := xjson.ParseString(jsonData)
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	fmt.Println("=== 新 API 演示 ===")

	// 1. 测试 MatchCount() - 总是返回匹配的节点数量
	fmt.Println("\n1. MatchCount() - 匹配节点数量:")

	booksResult := doc.Query("store.book")
	fmt.Printf("store.book 匹配数量: %d\n", booksResult.MatchCount())

	bookResult := doc.Query("store.book[0]")
	fmt.Printf("store.book[0] 匹配数量: %d\n", bookResult.MatchCount())

	priceResult := doc.Query("store.book[0].price")
	fmt.Printf("store.book[0].price 匹配数量: %d\n", priceResult.MatchCount())

	// 2. 测试 Size() - 返回单个匹配节点的大小
	fmt.Println("\n2. Size() - 单个节点大小:")

	booksSize, err := booksResult.Size()
	if err != nil {
		fmt.Printf("store.book Size() 错误: %v\n", err)
	} else {
		fmt.Printf("store.book 大小: %d\n", booksSize)
	}

	bookSize2, err := bookResult.Size()
	if err != nil {
		fmt.Printf("store.book[0] Size() 错误: %v\n", err)
	} else {
		fmt.Printf("store.book[0] 大小: %d\n", bookSize2)
	}

	priceSize, err := priceResult.Size()
	if err != nil {
		fmt.Printf("store.book[0].price Size() 错误: %v\n", err)
	} else {
		fmt.Printf("store.book[0].price 大小: %d\n", priceSize)
	}

	// 3. 测试 Value() 和 Values()
	fmt.Println("\n3. Value() 和 Values():")

	value, err := priceResult.Value()
	if err != nil {
		fmt.Printf("store.book[0].price Value() 错误: %v\n", err)
	} else {
		fmt.Printf("store.book[0].price 值: %v\n", value)
	}

	values := booksResult.Values()
	fmt.Printf("store.book 所有值数量: %d\n", len(values))

	// 4. 测试 Error()
	fmt.Println("\n4. Error() - 错误处理:")

	if booksResult.Error() != nil {
		fmt.Printf("store.book 错误: %v\n", booksResult.Error())
	} else {
		fmt.Println("store.book 查询成功")
	}

	// 5. 对比旧的 Count() 和新的 API
	fmt.Println("\n5. 对比旧 API 和新 API:")

	bicycleResult := doc.Query("store.bicycle")
	fmt.Printf("store.bicycle.Count() (旧 - 混淆行为): %d\n", bicycleResult.Count())
	fmt.Printf("store.bicycle.MatchCount() (新 - 匹配数量): %d\n", bicycleResult.MatchCount())
	bicycleSize, _ := bicycleResult.Size()
	fmt.Printf("store.bicycle.Size() (新 - 节点大小): %d\n", bicycleSize)

	// 6. 测试错误处理
	fmt.Println("\n6. 错误处理测试:")

	invalidResult := doc.Query("store.nonexistent")
	fmt.Printf("无效查询 MatchCount(): %d\n", invalidResult.MatchCount())
	invalidSize, _ := invalidResult.Size()
	fmt.Printf("无效查询 Size() 错误: %v\n", invalidSize)
	fmt.Printf("无效查询 Error(): %v\n", invalidResult.Error())

	fmt.Println("\n=== 总结 ===")
	fmt.Println("✅ MatchCount() - 清晰地返回匹配的节点数量")
	fmt.Println("✅ Size() - 清晰地返回单个节点的大小")
	fmt.Println("✅ Value() - 安全地获取单个值")
	fmt.Println("✅ Values() - 获取所有匹配值")
	fmt.Println("✅ Error() - 正确的错误处理")
	fmt.Println("✅ 职责分离 - 每个方法都有明确的职责")
}
