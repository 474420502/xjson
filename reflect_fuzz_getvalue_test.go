package xjson

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// generateRandomPath generates a random query path for fuzzing
func generateRandomPath(depth int) string {
	if depth <= 0 {
		return ""
	}
	// Simplified to generate valid keys for fuzzing Query
	return fmt.Sprintf("key%d", rand.Intn(100))
}

// FuzzGetValue tests the Query method with a variety of random inputs to catch panics
func FuzzGetValue(f *testing.F) {
	// Seed with some interesting values
	f.Add(`{"a":{"b":{"c":"d"}}}`, "/a/b/c")
	f.Add(`{"a":[1,2,3]}`, "/a[1]")
	f.Add(`{"a.b":"c"}`, "/a.b")
	f.Add(`{}`, "/")
	f.Add(`[]`, "[0]")

	rand.Seed(time.Now().UnixNano())

	f.Fuzz(func(t *testing.T, jsonData string, path string) {
		// Ensure jsonData is valid JSON, otherwise skip
		doc, err := ParseString(jsonData)
		if err != nil {
			return // Skip invalid JSON
		}

		// The goal of this fuzz test is to ensure that Query never panics,
		// regardless of the input path. We don't check the result, just that it completes.
		_ = doc.Query(path)

		// Also test with some randomly generated deep paths
		for i := 0; i < 5; i++ {
			randomPath := "/" + generateRandomPath(rand.Intn(5)+1)
			_ = doc.Query(randomPath)
		}
	})
}
