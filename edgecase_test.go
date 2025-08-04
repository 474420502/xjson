package xjson

import (
	"testing"
)

// 测试 Document.Set/Document.Delete 的 error 分支（只读Result）
func TestDocumentSetDeleteReadOnlyResult(t *testing.T) {
	doc := &Document{isValid: true, isMaterialized: true}
	doc.materialized = map[string]interface{}{"a": 1}
	doc.err = ErrReadOnlyResult
	if err := doc.Set("a", 2); err == nil {
		t.Error("只读Result Set 应报错")
	}
	if err := doc.Delete("a"); err == nil {
		t.Error("只读Result Delete 应报错")
	}
}

// 测试 Result.Get/Index/First/Last 的 error分支
func TestResultGetIndexFirstLastError(t *testing.T) {
	r := &Result{err: ErrTypeMismatch}
	if rr, ok := r.Get("a").(*Result); !ok || rr.err == nil {
		t.Error("Result.Get error分支应返回err")
	}
	if rr, ok := r.Index(0).(*Result); !ok || rr.err == nil {
		t.Error("Result.Index error分支应返回err")
	}
	if rr, ok := r.First().(*Result); !ok || rr.err == nil {
		t.Error("Result.First error分支应返回err")
	}
	if rr, ok := r.Last().(*Result); !ok || rr.err == nil {
		t.Error("Result.Last error分支应返回err")
	}
}

// 测试 Result.Keys/Count 的 error分支
func TestResultKeysCountError(t *testing.T) {
	r := &Result{err: ErrTypeMismatch}
	if r.Keys() != nil {
		t.Error("Result.Keys error分支应为nil")
	}
	if r.Count() != 0 {
		t.Error("Result.Count error分支应为0")
	}
}
