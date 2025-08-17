package xjson

import (
	"bytes"
	"encoding/json"
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
    "users": [
        {
            "id": 1,
            "name": "John Doe",
            "age": 30,
            "email": "john.doe@example.com",
            "isStudent": false,
            "courses": [
                {"title": "Math", "score": 90, "credits": 3},
                {"title": "Science", "score": 85, "credits": 4},
                {"title": "History", "score": 78, "credits": 3}
            ],
            "address": {
                "street": "123 Main St",
                "city": "Anytown",
                "zip": "12345",
                "coordinates": {
                    "latitude": 40.7128,
                    "longitude": -74.0060
                }
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
        },
        {
            "id": 2,
            "name": "Jane Smith",
            "age": 25,
            "email": "jane.smith@example.com",
            "isStudent": true,
            "courses": [
                {"title": "Physics", "score": 95, "credits": 4},
                {"title": "Chemistry", "score": 87, "credits": 3},
                {"title": "Biology", "score": 91, "credits": 4}
            ],
            "address": {
                "street": "456 Oak Ave",
                "city": "Another City",
                "zip": "67890",
                "coordinates": {
                    "latitude": 34.0522,
                    "longitude": -118.2437
                }
            },
            "hobbies": ["painting", "music", "travel"],
            "grades": [
                [92, 89, 94],
                [88, 90, 85]
            ],
            "notes": "Another long note for performance testing with a considerable amount of text to increase the size of the JSON data being processed in these benchmarks. This helps simulate real-world scenarios where JSON documents can be quite large and complex.",
            "metadata": {
                "id": "def-456",
                "timestamp": "2023-02-01T14:30:00Z",
                "tags": ["test", "performance", "benchmark", "student"]
            }
        },
        {
            "id": 3,
            "name": "Bob Johnson",
            "age": 35,
            "email": "bob.johnson@example.com",
            "isStudent": false,
            "courses": [
                {"title": "Literature", "score": 82, "credits": 3},
                {"title": "Philosophy", "score": 79, "credits": 3},
                {"title": "Art History", "score": 88, "credits": 4}
            ],
            "address": {
                "street": "789 Pine Rd",
                "city": "Yet Another City",
                "zip": "54321",
                "coordinates": {
                    "latitude": 41.8781,
                    "longitude": -87.6298
                }
            },
            "hobbies": ["cooking", "gardening", "photography"],
            "grades": [
                [85, 87, 80],
                [90, 85, 88]
            ],
            "notes": "Yet another extensive note to further increase the size of our test JSON data. This ensures that our benchmarks are working with realistically sized documents that can help us accurately measure performance differences between various JSON processing libraries and techniques.",
            "metadata": {
                "id": "ghi-789",
                "timestamp": "2023-03-01T10:15:00Z",
                "tags": ["test", "performance", "benchmark", "employee"]
            }
        },
        {
            "id": 4,
            "name": "Alice Williams",
            "age": 28,
            "email": "alice.williams@example.com",
            "isStudent": true,
            "courses": [
                {"title": "Computer Science", "score": 96, "credits": 4},
                {"title": "Mathematics", "score": 93, "credits": 3},
                {"title": "Statistics", "score": 89, "credits": 3}
            ],
            "address": {
                "street": "101 Elm St",
                "city": "Tech City",
                "zip": "98765",
                "coordinates": {
                    "latitude": 37.7749,
                    "longitude": -122.4194
                }
            },
            "hobbies": ["programming", "gaming", "reading"],
            "grades": [
                [94, 92, 95],
                [87, 90, 91]
            ],
            "notes": "More sample text to increase the overall size of the JSON document for more realistic benchmarking. The larger the document, the more apparent performance differences will be between different JSON processing approaches and libraries.",
            "metadata": {
                "id": "jkl-012",
                "timestamp": "2023-04-01T16:45:00Z",
                "tags": ["test", "performance", "benchmark", "student", "tech"]
            }
        },
        {
            "id": 5,
            "name": "Charlie Brown",
            "age": 32,
            "email": "charlie.brown@example.com",
            "isStudent": false,
            "courses": [
                {"title": "Marketing", "score": 85, "credits": 3},
                {"title": "Finance", "score": 88, "credits": 4},
                {"title": "Economics", "score": 82, "credits": 3}
            ],
            "address": {
                "street": "202 Maple Dr",
                "city": "Business City",
                "zip": "13579",
                "coordinates": {
                    "latitude": 29.7604,
                    "longitude": -95.3698
                }
            },
            "hobbies": ["investing", "tennis", "travel"],
            "grades": [
                [83, 85, 80],
                [86, 88, 84]
            ],
            "notes": "Additional text to further expand the size of the JSON data. This helps ensure that our benchmarks are working with substantial documents that can reveal performance characteristics that might not be apparent with smaller JSON structures.",
            "metadata": {
                "id": "mno-345",
                "timestamp": "2023-05-01T09:20:00Z",
                "tags": ["test", "performance", "benchmark", "employee", "business"]
            }
        },
        {
            "id": 6,
            "name": "Diana Prince",
            "age": 27,
            "email": "diana.prince@example.com",
            "isStudent": true,
            "courses": [
                {"title": "Psychology", "score": 94, "credits": 3},
                {"title": "Sociology", "score": 91, "credits": 3},
                {"title": "Anthropology", "score": 89, "credits": 4}
            ],
            "address": {
                "street": "303 Cedar Ln",
                "city": "Academic City",
                "zip": "24680",
                "coordinates": {
                    "latitude": 33.7490,
                    "longitude": -84.3880
                }
            },
            "hobbies": ["yoga", "meditation", "reading"],
            "grades": [
                [91, 89, 93],
                [88, 90, 87]
            ],
            "notes": "Even more sample text to make our test JSON document larger and more complex. This complexity helps us better understand how different JSON libraries perform under realistic conditions with nested structures and varied data types.",
            "metadata": {
                "id": "pqr-678",
                "timestamp": "2023-06-01T11:30:00Z",
                "tags": ["test", "performance", "benchmark", "student", "academic"]
            }
        },
        {
            "id": 7,
            "name": "Edward Norton",
            "age": 38,
            "email": "edward.norton@example.com",
            "isStudent": false,
            "courses": [
                {"title": "Engineering", "score": 87, "credits": 4},
                {"title": "Physics", "score": 85, "credits": 4},
                {"title": "Mathematics", "score": 83, "credits": 3}
            ],
            "address": {
                "street": "404 Birch St",
                "city": "Engineering City",
                "zip": "11223",
                "coordinates": {
                    "latitude": 39.9526,
                    "longitude": -75.1652
                }
            },
            "hobbies": ["building", "electronics", "robotics"],
            "grades": [
                [84, 86, 82],
                [89, 87, 90]
            ],
            "notes": "Additional content to increase the JSON document size for more accurate benchmarking. Larger documents help us identify performance bottlenecks and differences between various JSON processing approaches more effectively.",
            "metadata": {
                "id": "stu-901",
                "timestamp": "2023-07-01T13:45:00Z",
                "tags": ["test", "performance", "benchmark", "employee", "engineering"]
            }
        },
        {
            "id": 8,
            "name": "Fiona Gallagher",
            "age": 24,
            "email": "fiona.gallagher@example.com",
            "isStudent": true,
            "courses": [
                {"title": "Medicine", "score": 95, "credits": 5},
                {"title": "Biology", "score": 93, "credits": 4},
                {"title": "Chemistry", "score": 90, "credits": 4}
            ],
            "address": {
                "street": "505 Spruce Ave",
                "city": "Medical City",
                "zip": "33445",
                "coordinates": {
                    "latitude": 32.7765,
                    "longitude": -96.7970
                }
            },
            "hobbies": ["volunteering", "swimming", "reading"],
            "grades": [
                [93, 95, 92],
                [88, 90, 91]
            ],
            "notes": "More text to further expand the JSON document size. This helps ensure our benchmarks are working with substantial data that can reveal meaningful performance differences between JSON processing libraries and techniques.",
            "metadata": {
                "id": "vwx-234",
                "timestamp": "2023-08-01T15:15:00Z",
                "tags": ["test", "performance", "benchmark", "student", "medical"]
            }
        },
        {
            "id": 9,
            "name": "George Lucas",
            "age": 45,
            "email": "george.lucas@example.com",
            "isStudent": false,
            "courses": [
                {"title": "Film Studies", "score": 88, "credits": 3},
                {"title": "Creative Writing", "score": 86, "credits": 3},
                {"title": "Digital Media", "score": 90, "credits": 4}
            ],
            "address": {
                "street": "606 Redwood Rd",
                "city": "Creative City",
                "zip": "55667",
                "coordinates": {
                    "latitude": 36.1699,
                    "longitude": -115.1398
                }
            },
            "hobbies": ["filmmaking", "screenwriting", "photography"],
            "grades": [
                [86, 84, 88],
                [90, 92, 89]
            ],
            "notes": "Additional sample text to increase the JSON document size for more realistic benchmarking. The more complex and larger the document, the better we can evaluate the performance characteristics of different JSON processing libraries.",
            "metadata": {
                "id": "yza-567",
                "timestamp": "2023-09-01T17:30:00Z",
                "tags": ["test", "performance", "benchmark", "employee", "creative"]
            }
        },
        {
            "id": 10,
            "name": "Helen Keller",
            "age": 33,
            "email": "helen.keller@example.com",
            "isStudent": true,
            "courses": [
                {"title": "Literature", "score": 94, "credits": 3},
                {"title": "History", "score": 92, "credits": 3},
                {"title": "Languages", "score": 96, "credits": 4}
            ],
            "address": {
                "street": "707 Sequoia St",
                "city": "Educational City",
                "zip": "77889",
                "coordinates": {
                    "latitude": 42.3601,
                    "longitude": -71.0589
                }
            },
            "hobbies": ["reading (braille)", "speaking", "advocacy"],
            "grades": [
                [94, 93, 95],
                [90, 92, 94]
            ],
            "notes": "Final block of sample text to make our test JSON document significantly larger. This larger size helps us better evaluate performance differences between JSON processing libraries and techniques in scenarios that more closely resemble real-world usage patterns.",
            "metadata": {
                "id": "bcd-890",
                "timestamp": "2023-10-01T19:45:00Z",
                "tags": ["test", "performance", "benchmark", "student", "educational"]
            }
        }
    ],
    "summary": {
        "totalUsers": 10,
        "averageAge": 31.3,
        "studentCount": 5,
        "employeeCount": 5,
        "courseCount": 30,
        "hobbyCount": 30,
        "cities": [
            "Anytown",
            "Another City",
            "Yet Another City",
            "Tech City",
            "Business City",
            "Academic City",
            "Engineering City",
            "Medical City",
            "Creative City",
            "Educational City"
        ],
        "statistics": {
            "minAge": 24,
            "maxAge": 45,
            "ageDistribution": {
                "20-29": 4,
                "30-39": 4,
                "40-49": 2
            },
            "scoreRanges": {
                "90-100": 15,
                "80-89": 12,
                "70-79": 3
            }
        }
    }
}`
	largeJSONData = []byte(data)

}

// BenchmarkXJSONParse 衡量 xjson 的 JSON 解析性能
func BenchmarkXJSONParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = MustParse(largeJSONData)
	}
}

// BenchmarkXJSONQuery 衡量 xjson 的 JSON 查询性能
func BenchmarkXJSONQuery(b *testing.B) {
	doc, err := MustParse(largeJSONData)
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
	doc, err := MustParse(largeJSONData)
	if err != nil {
		b.Fatal(err)
	}
	setPath := "address/city"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.SetByPath(setPath, "NewCity")
		_ = doc.String()
	}
}

// 一次性解析后多次查询（预解析）
func BenchmarkXJSONQuery_OnceParse_MultiQuery(b *testing.B) {
	doc, err := MustParse(largeJSONData)
	if err != nil {
		b.Fatal(err)
	}
	queryPath := "address/city"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.Query(queryPath)
	}
}

// 每次懒解析+查询
func BenchmarkXJSONQuery_LazyParse_EachQuery(b *testing.B) {
	queryPath := "address/city"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, err := Parse(largeJSONData)
		if err != nil {
			b.Fatal(err)
		}
		doc.Query(queryPath)
	}
}

// BenchmarkGJSONQuery 衡量 gjson 的 JSON 查询性能
func BenchmarkGJSONQuery(b *testing.B) {
	queryPath := "address.city"
	for i := 0; i < b.N; i++ {
		gjson.Get(string(largeJSONData), queryPath)
	}
}

// gjson 一次性解析后多次查询（gjson 本身是懒解析，直接多次 Get）
func BenchmarkGJSONQuery_MultiQuery(b *testing.B) {
	queryPath := "address.city"
	b.ResetTimer()
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(largeJSONData, &data)
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

// json-iterator/go 一次性解析后多次查询
func BenchmarkJsonIterQuery_OnceParse_MultiQuery(b *testing.B) {
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

// json-iterator/go 每次懒解析+查询
func BenchmarkJsonIterQuery_LazyParse_EachQuery(b *testing.B) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
		json.Unmarshal(largeJSONData, &data)
		addr, ok := data["address"].(map[string]interface{})
		if ok {
			_ = addr["city"]
		}
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

// encoding/json 一次性解析后多次查询
func BenchmarkStandardJSONQuery_OnceParse_MultiQuery(b *testing.B) {
	var data map[string]interface{}
	json.Unmarshal(largeJSONData, &data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr, ok := data["address"].(map[string]interface{})
		if ok {
			_ = addr["city"]
		}
	}
}

// encoding/json 每次懒解析+查询
func BenchmarkStandardJSONQuery_LazyParse_EachQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
		json.Unmarshal(largeJSONData, &data)
		addr, ok := data["address"].(map[string]interface{})
		if ok {
			_ = addr["city"]
		}
	}
}
