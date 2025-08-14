package xjson

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

// largeJSONData 用于基准测试的大型 JSON 数据
var largeJSONData []byte

func init() {
	// 从一个文件加载大型 JSON 数据，或者直接定义一个大的字符串
	// 为了简化，这里先用一个简单的JSON，实际测试应替换为大型数据
	// 比如从 examples/data/large.json 读取
	// 或者直接定义一个足够大的 JSON 字符串
	data := `{
		"name": "John Doe",
		"age": 30,
		"isStudent": false,
		"courses": [
			{"title": "Math", "score": 90},
			{"title": "Science", "score": 85},
			{"title": "History", "score": 78}
		],
		"address": {
			"street": "123 Main St",
			"city": "Anytown",
			"zip": "12345"
		},
		"hobbies": ["reading", "hiking", "coding"],
		"grades": [
			[90, 88, 92],
			[75, 80, 82]
		],
		"notes": "This is a long note that will be parsed and re-parsed multiple times to simulate real-world scenarios where JSON data might be extensive and contain various data types. The goal is to measure the performance of different JSON parsing and querying libraries under varying conditions.",
		"metadata": {
			"id": "abc-123",
			"timestamp": "2023-01-01T12:00:00Z",
			"tags": ["test", "performance", "benchmark"]
		}
	}`
	largeJSONData = []byte(data)

	// 尝试从文件加载更复杂的 JSON 数据
	// 假设存在一个 large.json 文件在当前目录或测试数据目录
	if content, err := ioutil.ReadFile("testdata/large.json"); err == nil {
		largeJSONData = content
	} else if content, err := ioutil.ReadFile("example/data/large.json"); err == nil {
		largeJSONData = content
	} else {
		// 如果文件不存在，则使用默认的 largeJSONData
		// fmt.Println("Warning: Could not load large.json, using default data for benchmark.")
	}
}

// BenchmarkXJSONParse 衡量 xjson 的 JSON 解析性能
func BenchmarkXJSONParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse(largeJSONData)
	}
}

// BenchmarkXJSONQuery 衡量 xjson 的 JSON 查询性能
func BenchmarkXJSONQuery(b *testing.B) {
	doc, err := Parse(largeJSONData)
	if err != nil {
		b.Fatal(err)
	}
	queryPath := "address/city"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.Query(queryPath)
	}
}

// BenchmarkXJSONSet 衡量 xjson 的 Set 方法性能
func BenchmarkXJSONSet(b *testing.B) {
	doc, err := Parse(largeJSONData)
	if err != nil {
		b.Fatal(err)
	}
	setPath := "address.city"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.Set(setPath, "NewCity")
	}
}

// BenchmarkStandardJSONUnmarshal 衡量 encoding/json 的 JSON 反序列化性能
func BenchmarkStandardJSONUnmarshal(b *testing.B) {
	var data map[string]interface{}
	for i := 0; i < b.N; i++ {
		json.Unmarshal(largeJSONData, &data)
	}
}

// BenchmarkStandardJSONDecode 衡量 encoding/json 的 JSON 解码性能 (流式)
func BenchmarkStandardJSONDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decoder := json.NewDecoder(bytes.NewReader(largeJSONData))
		var data interface{}
		decoder.Decode(&data)
	}
}

// BenchmarkStandardJSONQuery 衡量 encoding/json 的查询性能
func BenchmarkStandardJSONQuery(b *testing.B) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
		json.Unmarshal(largeJSONData, &data)
		// 手动模拟路径查询
		val, ok := data["address"].(map[string]interface{})
		if ok {
			_ = val["city"]
		}
	}
}

// BenchmarkStandardJSONSet 衡量 encoding/json 的设置性能
func BenchmarkStandardJSONSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
		json.Unmarshal(largeJSONData, &data)
		addr, ok := data["address"].(map[string]interface{})
		if ok {
			addr["city"] = "NewCity"
		}
		_, _ = json.Marshal(data)
	}
}

// BenchmarkGJSONQuery 衡量 gjson 的 JSON 查询性能
func BenchmarkGJSONQuery(b *testing.B) {
	queryPath := "address.city"
	for i := 0; i < b.N; i++ {
		gjson.Get(string(largeJSONData), queryPath)
	}
}

// 引入 json-iterator/go

// BenchmarkJsonIterUnmarshal 衡量 json-iterator/go 的 JSON 反序列化性能
func BenchmarkJsonIterUnmarshal(b *testing.B) {
	var data map[string]interface{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	for i := 0; i < b.N; i++ {
		json.Unmarshal(largeJSONData, &data)
	}
}

// BenchmarkJsonIterQuery 衡量 json-iterator/go 的查询性能
func BenchmarkJsonIterQuery(b *testing.B) {
	var data map[string]interface{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	json.Unmarshal(largeJSONData, &data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr, ok := data["address"].(map[string]interface{})
		if ok {
			_ = addr["city"]
		}
	}
}

// BenchmarkJsonIterSet 衡量 json-iterator/go 的设置性能
func BenchmarkJsonIterSet(b *testing.B) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
		json.Unmarshal(largeJSONData, &data)
		addr, ok := data["address"].(map[string]interface{})
		if ok {
			addr["city"] = "NewCity"
		}
		_, _ = json.Marshal(data)
	}
}
