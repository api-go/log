package log_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"apigo.cc/go/log"
)

func TestSplitTag(t *testing.T) {
	logFile := "test_rotate.log"
	// 使用每秒切分的标签，方便测试文件轮转
	splitTag := ".20060102150405" 
	
	conf := log.Config{
		Name:     "test-split",
		Level:    "info",
		File:     logFile,
		SplitTag: splitTag,
	}
	logger := log.NewLogger(conf)
	
	// 1. 记录第一条日志
	logger.Info("first message")
	time.Sleep(300 * time.Millisecond) // 等待异步写入
	expectedFile1 := logFile + "." + time.Now().Format(splitTag)
	
	if _, err := os.Stat(expectedFile1); os.IsNotExist(err) {
		t.Fatalf("First log file %s does not exist", expectedFile1)
	}

	// 2. 等待跨秒
	time.Sleep(1100 * time.Millisecond)
	
	// 3. 记录第二条日志，触发轮转
	logger.Info("second message")
	time.Sleep(300 * time.Millisecond) // 等待异步写入
	expectedFile2 := logFile + "." + time.Now().Format(splitTag)
	
	if expectedFile1 == expectedFile2 {
		t.Errorf("Files should be different for rotation, but both are %s", expectedFile1)
	}

	if _, err := os.Stat(expectedFile2); os.IsNotExist(err) {
		t.Errorf("Second log file %s does not exist after rotation", expectedFile2)
	}

	// 清理
	os.Remove(expectedFile1)
	os.Remove(expectedFile2)
}

func TestSensitiveDetailed(t *testing.T) {
	type SecretLog struct {
		log.BaseLog
		Password  string
		SecretKey string
		SafeData  string
	}

	entry := log.GetEntry[SecretLog]()
	entry.BaseLog.LogType = "secret"
	entry.Password = "my_password"
	entry.SecretKey = "super_secret"
	entry.SafeData = "hello"
	
	// 直接测试 ToArrayBytes
	// 注意：passed to ToArrayBytes 的 keys 应该是已经过 fixField 处理的
	sensitiveKeys := []string{"password", "secretkey"}
	buf := log.ToArrayBytes(entry, sensitiveKeys)
	result := string(buf)
	
	if strings.Contains(result, "my_password") {
		t.Errorf("Sensitive data 'my_password' not masked in: %s", result)
	}
	if strings.Contains(result, "super_secret") {
		t.Errorf("Sensitive data 'super_secret' not masked in: %s", result)
	}
	if !strings.Contains(result, "hello") {
		t.Errorf("Safe data 'hello' should be present in: %s", result)
	}
}

func TestDeepDesensitization(t *testing.T) {
	type Nested struct {
		Password string
		Token    string
	}
	type DeepLog struct {
		log.BaseLog
		Data map[string]any
		User Nested
	}

	entry := log.GetEntry[DeepLog]()
	entry.BaseLog.LogType = "deep"
	entry.Data = map[string]any{
		"password": "data_password",
		"safe":     "data_safe",
	}
	entry.User = Nested{
		Password: "user_password",
		Token:    "user_token",
	}

	sensitiveKeys := []string{"password", "token"}
	buf := log.ToArrayBytes(entry, sensitiveKeys)
	result := string(buf)

	// Check deep desensitization in map
	if strings.Contains(result, "data_password") {
		t.Errorf("Nested map data 'data_password' not masked in: %s", result)
	}
	// Check deep desensitization in struct
	if strings.Contains(result, "user_password") {
		t.Errorf("Nested struct data 'user_password' not masked in: %s", result)
	}
	if strings.Contains(result, "user_token") {
		t.Errorf("Nested struct data 'user_token' not masked in: %s", result)
	}
	if !strings.Contains(result, "data_safe") {
		t.Errorf("Safe data 'data_safe' should be present in: %s", result)
	}
}

