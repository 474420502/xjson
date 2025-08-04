package xjson

import (
	"testing"
)

// 测试 materialize 的 json.Unmarshal 失败分支
func TestLowCoverageFunctions_Document_materialize(t *testing.T) {
	d := &Document{isValid: true, raw: []byte("{invalid json}")}
	err := d.materialize()
	if err == nil {
		t.Error("materialize 非法json应报错")
	}
}

// 测试 Result.Bytes 的 error 分支
func TestLowCoverageFunctions_Result_Bytes(t *testing.T) {
	r := &Result{err: ErrTypeMismatch}
	b, err := r.Bytes()
	if err == nil || b != nil {
		t.Error("Result.Bytes error分支应返回err且b为nil")
	}

	r = &Result{}
	// 空matches
	b, err = r.Bytes()
	if err == nil || b != nil {
		t.Error("Result.Bytes 空matches应返回err且b为nil")
	}
}
