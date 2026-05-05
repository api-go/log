package log_test

import (
	"testing"

	"apigo.cc/go/log"
)

// Define local log types for testing since they are commented out in the main package
type RequestEntry struct {
	log.BaseLog
	Method       string
	Path         string
	ResponseCode int
	UsedTime     float32
}

type DBEntry struct {
	log.BaseLog
	DbType    string
	Dsn       string
	Query     string
	QueryArgs string
	UsedTime  float32
	Error     string
}

func TestLogger(t *testing.T) {
	conf := log.Config{
		Name:  "test-app",
		Level: "debug",
	}
	logger := log.NewLogger(conf)

	// 测试 Info 日志
	logger.Info("hello", "key1", "value1")
}

func TestDesensitization(t *testing.T) {
	logger := log.NewLogger(log.Config{
		Sensitive: "phone",
	})

	type MyLog struct {
		log.BaseLog
		Phone string
	}

	entry := log.GetEntry[MyLog]()
	entry.Phone = "13812345678"
	logger.Log(entry) // 应该在输出中脱敏
}

func TestDBLog(t *testing.T) {
	logger := log.NewLogger(log.Config{
		Level: "debug",
	})

	entry := log.GetEntry[DBEntry]()
	entry.LogType = "db"
	entry.DbType = "mysql"
	entry.Query = "SELECT * FROM users"
	entry.UsedTime = 10.5
	logger.Log(entry)
}

func TestRequestLog(t *testing.T) {
	logger := log.NewLogger(log.Config{
		Level: "debug",
	})

	entry := log.GetEntry[RequestEntry]()
	entry.LogType = "request"
	entry.Method = "GET"
	entry.Path = "/api/user"
	entry.ResponseCode = 200
	entry.UsedTime = 10.5
	logger.Log(entry)
}

func TestExtraLogs(t *testing.T) {
	logger := log.NewLogger(log.Config{
		Level: "debug",
	})

	logger.Info("Extra log test", "key", "value")
}
