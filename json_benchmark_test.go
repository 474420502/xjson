package xjson

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/474420502/xjson/internal/engine"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

// largeJSONData 用于基准测试的大型 JSON 数据
var largeJSONData []byte

const (
  xjsonQueryPath  = "/level1/level2/level3/level4/level5/level6/level7/level8/level9/level10/users[0]/profile/personal/name"
  xjsonSetPath    = "/level1/level2/level3/level4/level5/level6/level7/level8/level9/level10/users[0]/profile/personal"
  gjsonQueryPath  = "level1.level2.level3.level4.level5.level6.level7.level8.level9.level10.users.0.profile.personal.name"
  updatedUserAge  = 31.0
)

var benchmarkJSON = jsoniter.ConfigCompatibleWithStandardLibrary

var benchmarkQuerySink any
var benchmarkBytesSink []byte
var benchmarkStringSink string

func benchmarkAgeInt(i int) int {
  if i&1 == 0 {
    return 31
  }
  return 32
}

func benchmarkAgeFloat(i int) float64 {
  if i&1 == 0 {
    return 31.0
  }
  return 32.0
}

func personalMapFromDecoded(data map[string]interface{}) map[string]interface{} {
  level1, ok := data["level1"].(map[string]interface{})
  if !ok {
    return nil
  }
  level2, ok := level1["level2"].(map[string]interface{})
  if !ok {
    return nil
  }
  level3, ok := level2["level3"].(map[string]interface{})
  if !ok {
    return nil
  }
  level4, ok := level3["level4"].(map[string]interface{})
  if !ok {
    return nil
  }
  level5, ok := level4["level5"].(map[string]interface{})
  if !ok {
    return nil
  }
  level6, ok := level5["level6"].(map[string]interface{})
  if !ok {
    return nil
  }
  level7, ok := level6["level7"].(map[string]interface{})
  if !ok {
    return nil
  }
  level8, ok := level7["level8"].(map[string]interface{})
  if !ok {
    return nil
  }
  level9, ok := level8["level9"].(map[string]interface{})
  if !ok {
    return nil
  }
  level10, ok := level9["level10"].(map[string]interface{})
  if !ok {
    return nil
  }
  users, ok := level10["users"].([]interface{})
  if !ok || len(users) == 0 {
    return nil
  }
  user0, ok := users[0].(map[string]interface{})
  if !ok {
    return nil
  }
  profile, ok := user0["profile"].(map[string]interface{})
  if !ok {
    return nil
  }
  personal, ok := profile["personal"].(map[string]interface{})
  if !ok {
    return nil
  }
  return personal
}

func nameFromDecoded(data map[string]interface{}) string {
  personal := personalMapFromDecoded(data)
  if personal == nil {
    return ""
  }
  name, _ := personal["name"].(string)
  return name
}

func init() {
	// 使用深度嵌套的 JSON 结构作为基准测试数据
	largeJSONData = []byte(`{
  "level1": {
    "level2": {
      "level3": {
        "level4": {
          "level5": {
            "level6": {
              "level7": {
                "level8": {
                  "level9": {
                    "level10": {
                      "data": "This is a deeply nested JSON structure for performance testing",
                      "users": [
                        {
                          "id": 1,
                          "profile": {
                            "personal": {
                              "name": "John Doe",
                              "age": 30,
                              "contact": {
                                "email": "john.doe@example.com",
                                "address": {
                                  "home": {
                                    "street": "123 Main St",
                                    "city": "Anytown",
                                    "state": "CA",
                                    "zip": "12345",
                                    "coordinates": {
                                      "latitude": 40.7128,
                                      "longitude": -74.0060,
                                      "elevation": {
                                        "meters": 10,
                                        "accuracy": {
                                          "horizontal": 0.5,
                                          "vertical": 1.2
                                        }
                                      }
                                    }
                                  },
                                  "work": {
                                    "street": "456 Office Blvd",
                                    "city": "Business City",
                                    "state": "CA",
                                    "zip": "67890"
                                  }
                                }
                              }
                            },
                            "education": {
                              "degrees": [
                                {
                                  "type": "Bachelor",
                                  "field": "Computer Science",
                                  "institution": {
                                    "name": "Tech University",
                                    "location": {
                                      "campus": "Main",
                                      "address": {
                                        "street": "University Ave",
                                        "city": "College Town",
                                        "state": "CA",
                                        "zip": "11111"
                                      }
                                    },
                                    "rankings": {
                                      "national": 50,
                                      "international": 200,
                                      "byField": {
                                        "computerScience": 25,
                                        "engineering": 30
                                      }
                                    }
                                  },
                                  "year": 2010,
                                  "gpa": 3.8
                                },
                                {
                                  "type": "Master",
                                  "field": "Software Engineering",
                                  "institution": {
                                    "name": "Graduate Institute",
                                    "location": {
                                      "campus": "North",
                                      "address": {
                                        "street": "Institute Road",
                                        "city": "Graduate City",
                                        "state": "CA",
                                        "zip": "22222"
                                      }
                                    },
                                    "rankings": {
                                      "national": 20,
                                      "international": 100,
                                      "byField": {
                                        "softwareEngineering": 5,
                                        "computerScience": 10
                                      }
                                    }
                                  },
                                  "year": 2012,
                                  "gpa": 3.9
                                }
                              ],
                              "courses": {
                                "undergraduate": {
                                  "core": [
                                    {
                                      "name": "Data Structures",
                                      "details": {
                                        "code": "CS101",
                                        "credits": 3,
                                        "instructor": {
                                          "name": "Dr. Smith",
                                          "department": {
                                            "name": "Computer Science",
                                            "faculty": {
                                              "name": "Engineering",
                                              "dean": {
                                                "name": "Dr. Johnson",
                                                "contact": {
                                                  "email": "dean@eng.univ.edu",
                                                  "phone": "123-456-7890"
                                                }
                                              }
                                            }
                                          }
                                        },
                                        "schedule": {
                                          "days": ["Monday", "Wednesday", "Friday"],
                                          "time": {
                                            "start": "09:00",
                                            "end": "09:50",
                                            "duration": 50
                                          }
                                        }
                                      }
                                    }
                                  ]
                                }
                              }
                            }
                          },
                          "employment": {
                            "current": {
                              "company": {
                                "name": "Tech Corp",
                                "address": {
                                  "headquarters": {
                                    "street": "Corp Plaza",
                                    "city": "Tech City",
                                    "state": "CA",
                                    "zip": "33333",
                                    "buildings": {
                                      "main": {
                                        "floors": 20,
                                        "offices": {
                                          "engineering": {
                                            "floor": 10,
                                            "room": "10A",
                                            "occupants": [
                                              {
                                                "name": "John Doe",
                                                "position": "Senior Developer",
                                                "team": {
                                                  "name": "Backend Team",
                                                  "lead": {
                                                    "name": "Team Lead",
                                                    "reportsTo": {
                                                      "name": "Engineering Manager",
                                                      "department": {
                                                        "name": "Engineering",
                                                        "director": {
                                                          "name": "Director of Engineering",
                                                          "reportsTo": {
                                                            "name": "CTO",
                                                            "executiveTeam": {
                                                              "members": [
                                                                "CEO",
                                                                "CTO",
                                                                "CFO",
                                                                "COO"
                                                              ]
                                                            }
                                                          }
                                                        }
                                                      }
                                                    }
                                                  }
                                                }
                                              }
                                            ]
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      ]
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`)
}

// BenchmarkXJSONParse 衡量 xjson 的 JSON 解析性能
func BenchmarkXJSONParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
    benchmarkQuerySink, _ = MustParse(largeJSONData)
	}
}

// BenchmarkXJSONQuery 衡量 xjson 的 JSON 查询性能
func BenchmarkXJSONQuery(b *testing.B) {
	doc, err := MustParse(largeJSONData)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
    benchmarkStringSink = doc.Query(xjsonQueryPath).String()
	}
}

func BenchmarkXJSONQuery_OnceParse_FirstHit(b *testing.B) {
  doc, err := MustParse(largeJSONData)
  if err != nil {
    b.Fatal(err)
  }
  inner := doc.(nodeWrapper).Node

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    engine.ResetQueryCache(inner)
    benchmarkStringSink = doc.Query(xjsonQueryPath).String()
  }
}

// BenchmarkXJSONSet 衡量 xjson 的 Set 方法性能
func BenchmarkXJSONSet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, err := MustParse(largeJSONData)
		if err != nil {
			b.Fatal(err)
		}
    result := doc.Query(xjsonSetPath)
		result.Set("age", 31)
    benchmarkQuerySink = doc.String()
	}
}

func BenchmarkXJSONSet_Prepared_MutateOnly(b *testing.B) {
  doc, err := MustParse(largeJSONData)
  if err != nil {
    b.Fatal(err)
  }
  target := doc.Query(xjsonSetPath)
  if err := target.Error(); err != nil {
    b.Fatal(err)
  }

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    benchmarkQuerySink = target.Set("age", benchmarkAgeInt(i))
  }
}

// 一次性解析后多次查询（预解析）
func BenchmarkXJSONQuery_OnceParse_MultiQuery(b *testing.B) {
	doc, err := MustParse(largeJSONData)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
    benchmarkStringSink = doc.Query(xjsonQueryPath).String()
	}
}

// 每次懒解析+查询
func BenchmarkXJSONQuery_LazyParse_EachQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, err := Parse(largeJSONData)
		if err != nil {
			b.Fatal(err)
		}
    benchmarkStringSink = doc.Query(xjsonQueryPath).String()
	}

}

func BenchmarkJsonIterParse(b *testing.B) {
  for i := 0; i < b.N; i++ {
    var data map[string]interface{}
    if err := benchmarkJSON.Unmarshal(largeJSONData, &data); err != nil {
      b.Fatal(err)
    }
    benchmarkQuerySink = data
  }
}

func BenchmarkStandardJSONParse(b *testing.B) {
  for i := 0; i < b.N; i++ {
    var data map[string]interface{}
    if err := json.Unmarshal(largeJSONData, &data); err != nil {
      b.Fatal(err)
    }
    benchmarkQuerySink = data
  }
}

// BenchmarkGJSONQuery 衡量 gjson 的 JSON 查询性能
func BenchmarkGJSONQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
    benchmarkStringSink = gjson.GetBytes(largeJSONData, gjsonQueryPath).String()
	}
}

// gjson 一次性解析后多次查询（gjson 本身是懒解析，直接多次 Get）
func BenchmarkGJSONQuery_MultiQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
    benchmarkStringSink = gjson.GetBytes(largeJSONData, gjsonQueryPath).String()
	}
}

// BenchmarkJsonIterQuery 衡量 json-iterator/go 的查询性能
func BenchmarkJsonIterQuery(b *testing.B) {
	var data map[string]interface{}
  if err := benchmarkJSON.Unmarshal(largeJSONData, &data); err != nil {
    b.Fatal(err)
  }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
    benchmarkStringSink = nameFromDecoded(data)
	}
}

// BenchmarkJsonIterSet 衡量 json-iterator/go 的设置性能
func BenchmarkJsonIterSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
    if err := benchmarkJSON.Unmarshal(largeJSONData, &data); err != nil {
      b.Fatal(err)
		}
    personal := personalMapFromDecoded(data)
    if personal == nil {
      b.Fatal("personal path not found")
    }
    personal["age"] = updatedUserAge
    out, err := benchmarkJSON.Marshal(data)
    if err != nil {
      b.Fatal(err)
    }
    benchmarkBytesSink = out
	}
}

func BenchmarkJsonIterSet_Prepared_MutateOnly(b *testing.B) {
  var data map[string]interface{}
  if err := benchmarkJSON.Unmarshal(largeJSONData, &data); err != nil {
    b.Fatal(err)
  }
  personal := personalMapFromDecoded(data)
  if personal == nil {
    b.Fatal("personal path not found")
  }

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    personal["age"] = benchmarkAgeFloat(i)
    benchmarkQuerySink = personal["age"]
  }
}

// json-iterator/go 一次性解析后多次查询
func BenchmarkJsonIterQuery_OnceParse_MultiQuery(b *testing.B) {
	var data map[string]interface{}
  if err := benchmarkJSON.Unmarshal(largeJSONData, &data); err != nil {
    b.Fatal(err)
  }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
    benchmarkStringSink = nameFromDecoded(data)
	}
}

// json-iterator/go 每次懒解析+查询
func BenchmarkJsonIterQuery_LazyParse_EachQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
    if err := benchmarkJSON.Unmarshal(largeJSONData, &data); err != nil {
      b.Fatal(err)
		}
    benchmarkStringSink = nameFromDecoded(data)
	}
}

// BenchmarkStandardJSONDecode 衡量 encoding/json 的 JSON 解码性能 (流式)
func BenchmarkStandardJSONDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decoder := json.NewDecoder(bytes.NewReader(largeJSONData))
		var data interface{}
    if err := decoder.Decode(&data); err != nil {
      b.Fatal(err)
    }
    benchmarkQuerySink = data
	}
}

// BenchmarkStandardJSONQuery 衡量 encoding/json 的查询性能
func BenchmarkStandardJSONQuery(b *testing.B) {
  var data map[string]interface{}
  if err := json.Unmarshal(largeJSONData, &data); err != nil {
    b.Fatal(err)
  }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
    benchmarkStringSink = nameFromDecoded(data)
	}
}

// BenchmarkStandardJSONSet 衡量 encoding/json 的设置性能
func BenchmarkStandardJSONSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
    if err := json.Unmarshal(largeJSONData, &data); err != nil {
      b.Fatal(err)
		}
    personal := personalMapFromDecoded(data)
    if personal == nil {
      b.Fatal("personal path not found")
    }
    personal["age"] = updatedUserAge
    out, err := json.Marshal(data)
    if err != nil {
      b.Fatal(err)
    }
    benchmarkBytesSink = out
	}
}

func BenchmarkStandardJSONSet_Prepared_MutateOnly(b *testing.B) {
  var data map[string]interface{}
  if err := json.Unmarshal(largeJSONData, &data); err != nil {
    b.Fatal(err)
  }
  personal := personalMapFromDecoded(data)
  if personal == nil {
    b.Fatal("personal path not found")
  }

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    personal["age"] = benchmarkAgeFloat(i)
    benchmarkQuerySink = personal["age"]
  }
}

// encoding/json 一次性解析后多次查询
func BenchmarkStandardJSONQuery_OnceParse_MultiQuery(b *testing.B) {
	var data map[string]interface{}
  if err := json.Unmarshal(largeJSONData, &data); err != nil {
    b.Fatal(err)
  }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
    benchmarkQuerySink = nameFromDecoded(data)
	}
}

// encoding/json 每次懒解析+查询
func BenchmarkStandardJSONQuery_LazyParse_EachQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
    if err := json.Unmarshal(largeJSONData, &data); err != nil {
      b.Fatal(err)
		}
    benchmarkStringSink = nameFromDecoded(data)
	}
}
