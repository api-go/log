package log

import (
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	conf := Config{
		Name:  "test-app",
		Level: "debug",
	}
	logger := NewLogger(conf)

	// 测试 Info 日志
	logger.Info("hello", "key1", "value1")
}

func TestDesensitization(t *testing.T) {
	logger := NewLogger(Config{
		Sensitive: "phone",
	})
	
	data := map[string]any{
		"phone": "13812345678",
	}
	logger.Log(data) // 应该在输出中脱敏
}

func TestDBLog(t *testing.T) {
	logger := NewLogger(Config{
		Level: "debug",
	})

	// 测试普通 DB 日志
	logger.DB("mysql", "dsn...", "SELECT * FROM users", []any{1}, 10.5, nil)

	// 测试 DB 错误日志 (通过传递 error 对象)
	logger.DB("mysql", "dsn...", "SELECT * FROM users", []any{1}, 10.5, fmt.Errorf("connection lost"))

	// 测试带额外参数的 DB 日志
	logger.DB("mysql", "dsn...", "SELECT * FROM users", []any{1}, 10.5, nil, "k1", "v1")
}

func TestRequestLog(t *testing.T) {
	logger := NewLogger(Config{
		Level: "debug",
	})

	req := &RequestLog{
		Method: "GET",
		Path:   "/api/user",
	}
	logger.Request(req)
}

func TestExtraLogs(t *testing.T) {
	logger := NewLogger(Config{
		Level: "debug",
	})

	logger.Task("CleanCache", 150.2, true, "Success clean")
	logger.Monitor("CPU", 1, "Normal")
	logger.Statistic("Business", "OrderCount", 100)
}
