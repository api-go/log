package log

import (
	"reflect"
	"testing"
)

// MockRequestLog 用于测试池化逻辑
type MockRequestLog struct {
	BaseLog
	RequestId string
	UsedTime  float32
}

func TestWithEntry(t *testing.T) {
	typ := reflect.TypeOf(&MockRequestLog{})

	WithEntry(typ, func(e any) {
		entry := e.(*MockRequestLog)
		entry.RequestId = "with-entry-id"
	})

	// 验证 PutEntry 自动被调用
	entry2 := GetEntry(typ).(*MockRequestLog)
	if entry2.RequestId != "" {
		t.Errorf("Expected reset, got %s", entry2.RequestId)
	}
}
