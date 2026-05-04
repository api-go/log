package log

import (
	"reflect"
	"testing"
)

// Field 定义单个日志字段
type Field struct {
	Key   string
	Value interface{}
}

// BenchRequestLog 作为测试用例，使用固定大小字段数组
type BenchRequestLog struct {
	RequestId string
	UsedTime  float32
	// 预留 10 个固定 Extra 字段位
	ExtraCount int
	Extra      [10]Field
}

func (r *BenchRequestLog) Reset() {
	r.RequestId = ""
	r.UsedTime = 0
	r.ExtraCount = 0
	// 结构体数组字段会自动清空，无需特殊处理
}

func BenchmarkLogger_RequestLog_Realistic(b *testing.B) {
	typ := reflect.TypeOf(&RequestLog{})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		WithEntry(typ, func(e any) {
			entry := e.(*RequestLog)
			entry.RequestId = "req-1234567890"
			entry.UsedTime = 45.67
			entry.Path = "/api/v1/user/profile"
			entry.Method = "POST"
			entry.ResponseCode = 200

			entry.RequestHeaders["Content-Type"] = "application/json"
			entry.RequestHeaders["Authorization"] = "Bearer token-value"

			entry.RequestData["userId"] = 10086
			entry.RequestData["action"] = "update_profile"

			entry.ResponseData = `{"status":"ok"}`
		})
	}
}
