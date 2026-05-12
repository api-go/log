package log_test

import (
	"strconv"
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
	logger.FillBase(&entry.BaseLog, "test")
	entry.Phone = "13812345678"
	logger.Log(entry) // 应该在输出中脱敏
}

func TestDBLog(t *testing.T) {
	logger := log.NewLogger(log.Config{
		Level: "debug",
	})

	entry := log.GetEntry[DBEntry]()
	logger.FillBase(&entry.BaseLog, "db")
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
	logger.FillBase(&entry.BaseLog, "request")
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

func TestAs(t *testing.T) {
	// 1. 测试 log.As (使用 DefaultLogger)
	val1 := log.As(strconv.Atoi("123"))
	if val1 != 123 {
		t.Errorf("log.As expected 123, got %v", val1)
	}

	val2 := log.As(strconv.Atoi("abc"))
	if val2 != 0 {
		t.Errorf("log.As expected 0, got %v", val2)
	}

	// 2. 测试 logger.As (方法)
	logger := log.NewLogger(log.Config{Level: "debug"})
	val3 := logger.As(strconv.Atoi("456")).(int)
	if val3 != 456 {
		t.Errorf("logger.As expected 456, got %v", val3)
	}

	val4 := logger.As(strconv.Atoi("def")).(int)
	if val4 != 0 {
		t.Errorf("logger.As expected 0, got %v", val4)
	}
}
