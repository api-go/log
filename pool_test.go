package log

import (
	"testing"
)

// MockRequestLog 用于测试池化逻辑
type MockRequestLog struct {
	BaseLog
	RequestId string
	UsedTime  float32
}

func (l *MockRequestLog) Reset() {
	l.BaseLog.Reset()
	l.RequestId = ""
	l.UsedTime = 0
}

func TestWithEntry(t *testing.T) {
	WithEntry(func(entry *MockRequestLog) {
		entry.RequestId = "with-entry-id"
	})

	// 验证 PutEntry 自动被调用
	entry2 := GetEntry[MockRequestLog]()
	if entry2.RequestId != "" {
		t.Errorf("Expected reset, got %s", entry2.RequestId)
	}
}
