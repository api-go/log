package log_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"apigo.cc/go/log"
)

func TestSplitTag(t *testing.T) {
	logFile := "test_split.log"
	// 使用每分钟切分的标签，方便测试 (当然实际可能需要模拟时间，但这里我们可以直接写然后检查文件名)
	splitTag := ".200601021504" 
	
	conf := log.Config{
		Name:     "test-split",
		Level:    "info",
		File:     logFile,
		SplitTag: splitTag,
	}
	logger := log.NewLogger(conf)
	
	logger.Info("split test message")
	
	// 给异步写入一点时间
	time.Sleep(300 * time.Millisecond)
	
	// 预期文件名
	expectedFile := logFile + "." + time.Now().Format(splitTag)
	
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected log file %s does not exist", expectedFile)
	} else {
		// 清理
		os.Remove(expectedFile)
	}
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
