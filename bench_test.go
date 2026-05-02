package log

import (
	"os"
	"testing"
)

func BenchmarkLoggerInfo(b *testing.B) {
	// 创建一个临时日志文件用于测试
	logger := NewLogger(Config{
		File: "bench.log",
	})
	defer os.Remove("bench.log")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("bench log", "index", i)
	}
	b.StopTimer()
}

func BenchmarkLoggerAsyncConcurrent(b *testing.B) {
	logger := NewLogger(Config{
		File: "bench_async.log",
	})
	defer os.Remove("bench_async.log")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Info("concurrent async log", "index", i)
			i++
		}
	})
	b.StopTimer()
}
