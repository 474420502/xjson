package engine

import "testing"

func TestSIMDOptimizedHelpers(t *testing.T) {
	t.Run("find quote optimized", func(t *testing.T) {
		small := []byte(`"hello"`)
		if got := findQuoteOptimized(small, 0); got != len(small)-1 {
			t.Fatalf("unexpected small quote index: %d", got)
		}

		large := append([]byte{'"'}, make([]byte, 120)...)
		large[len(large)-1] = '"'
		if got := findQuoteOptimized(large, 0); got != len(large)-1 {
			t.Fatalf("unexpected large quote index: %d", got)
		}
		if got := findQuoteOptimized([]byte("abc"), 0); got != -1 {
			t.Fatalf("expected -1, got %d", got)
		}
	})

	t.Run("find quote implementations with escapes", func(t *testing.T) {
		data := []byte{'"', 'a', '\\', '\\', '"', 'b', '"'}
		if got := findQuoteSimple(data, 0); got != 4 {
			t.Fatalf("findQuoteSimple got %d", got)
		}

		large := append([]byte{'"'}, make([]byte, 80)...)
		large[len(large)-1] = '"'
		if got := findQuoteWord(large, 0); got != len(large)-1 {
			t.Fatalf("findQuoteWord got %d", got)
		}
	})

	t.Run("find brace optimized", func(t *testing.T) {
		data := []byte(`{"a":{"b":"{x}"},"c":1}`)
		if got := findBraceOptimized(data, 0); got != len(data)-1 {
			t.Fatalf("unexpected brace end: %d", got)
		}
		if got := findBraceOptimized([]byte("[]"), 0); got != -1 {
			t.Fatalf("expected -1, got %d", got)
		}
	})

	t.Run("find bracket optimized", func(t *testing.T) {
		data := []byte(`[1,[2,3],"[]",{"a":1}]`)
		if got := findBracketOptimized(data, 0); got != len(data)-1 {
			t.Fatalf("unexpected bracket end: %d", got)
		}
		if got := findBracketOptimized([]byte("{}"), 0); got != -1 {
			t.Fatalf("expected -1, got %d", got)
		}
	})

	t.Run("has zero byte", func(t *testing.T) {
		if !hasZeroByte(0x1100223344556677) {
			t.Fatal("expected zero byte detection")
		}
		if hasZeroByte(0x1122334455667788) {
			t.Fatal("did not expect zero byte detection")
		}
	})
}