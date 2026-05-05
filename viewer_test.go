package log_test

import (
	"strings"
	"testing"

	"apigo.cc/go/log"
)

func TestViewable(t *testing.T) {
	// First ensure mock_info type is registered so we have meta
	entry := &log.InfoLog{
		BaseLog: log.BaseLog{
			LogName: "test-app",
			LogType: "info",
		},
		Info: "hello world",
	}
	log.RegisterType("info", entry)

	line := `["test-app","info",1714896000000000000,"trace-123","","","","","hello world",{"key":"value"}]`
	out := log.Viewable(line)

	if !strings.Contains(out, "hello world") {
		t.Errorf("expected 'hello world' in output, got: %s", out)
	}
	if !strings.Contains(out, "trace-123") {
		t.Errorf("expected 'trace-123' in output, got: %s", out)
	}
	if !strings.Contains(out, "key:") {
		t.Errorf("expected 'key:' in output, got: %s", out)
	}
	if !strings.Contains(out, "value") {
		t.Errorf("expected 'value' in output, got: %s", out)
	}
}

func BenchmarkViewable(b *testing.B) {
	line := `["test-app","info",1714896000000000000,"trace-123","","","","","hello world",{"key":"value"}]`
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = log.Viewable(line)
	}
}
