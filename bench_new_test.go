package log_test

import (
	"testing"

	"github.com/api-go/log"
)

// Define local log types for testing since they are commented out in the main package
type BenchRequestLog struct {
	log.BaseLog
	Method       string
	Path         string
	ResponseCode int
	UsedTime     float32
}

func BenchmarkLogger_RequestLog_Realistic(b *testing.B) {
	// 使用文件模式但指向空设备，或者在配置中支持 Discard
	conf := log.Config{NoCompress: true, Level: "info", File: "/dev/null"}
	logger := log.NewLogger(conf)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		entry := log.GetEntry[BenchRequestLog]()
		// ... 填充逻辑 ...
		entry.LogType = "request"
		entry.Method = "POST"
		entry.Path = "/api/v1/user/profile"
		entry.ResponseCode = 200
		entry.UsedTime = 45.67
		logger.Log(entry)
	}
}
