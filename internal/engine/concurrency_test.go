package engine

import (
	"sync"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
)

// TestLazyParse_NoDeadlock_ConcurrentAccess tries to reproduce past deadlocks
// by concurrently accessing lazily-parsed nodes via Get, Index and Query.
func TestLazyParse_NoDeadlock_ConcurrentAccess(t *testing.T) {
	// Construct a reasonably nested JSON to exercise object/array lazy parsing
	jsonData := []byte(`{
        "a": [
            {"b": {"c": [1,2,3,{"d":"e"}]}},
            {"b": {"c": [4,5,6]}},
            {"b": {"c": [{"d":"x"}]}}
        ],
        "x": {"y": {"z":"value"}},
        "m": [0,1,2,3,4,5,6,7,8,9]
    }`)

	root, err := Parse(jsonData)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Ensure root is object
	if root.Type() != core.Object {
		t.Fatalf("expected object root, got %v", root.Type())
	}

	// Run many goroutines performing mixed operations
	var wg sync.WaitGroup
	goroutines := 50
	iterations := 200

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Query deep path
				q := root.Query("/a[0]/b/c[3]/d")
				if q != nil && q.IsValid() {
					_ = q.String()
				}

				// Alternate queries
				q2 := root.Query("/x/y/z")
				if q2 != nil && q2.IsValid() {
					_ = q2.String()
				}

				// Access via Get then Index when appropriate
				a := root.Get("a")
				if a != nil && a.IsValid() && a.Type() == core.Array {
					_ = a.Index(0).String()
				}

				// Access an index in m to exercise array index parsing
				m := root.Get("m")
				if m != nil && m.IsValid() && m.Type() == core.Array {
					_ = m.Index((j % 10)).String()
				}
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(8 * time.Second):
		t.Fatal("concurrent lazy-parse test timed out — possible deadlock")
	}
}
