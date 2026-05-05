package log_test

import (
	"testing"

	"apigo.cc/go/log"
)

func BenchmarkViewable(b *testing.B) {
	// 准备一个典型的 JSON 日志行，注意 Info, Warning 等在顶层
	line := `{"LogName":"test-app","LogType":"info","LogTime":1714896000000000000,"TraceId":"trace-123","info":"hello world","Extra":{"key":"value"}}`
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = log.Viewable(line)
	}
}

func BenchmarkViewable_Request(b *testing.B) {
	// RequestLog 的字段也在顶层
	line := `{"LogName":"test-app","LogType":"request","LogTime":1714896000000000000,"TraceId":"trace-123","method":"GET","path":"/api/user","responsecode":200,"usedtime":10.5}`
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = log.Viewable(line)
	}
}
