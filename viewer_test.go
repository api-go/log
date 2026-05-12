package log_test

import (
	"os"
	"strings"
	"testing"

	"apigo.cc/go/log"
)

func TestViewable(t *testing.T) {
	entry := &log.InfoLog{
		BaseLog: log.BaseLog{
			LogName: "test-app",
			LogType: "info",
			LogTime: 1714896000000000000,
			TraceId: "trace-123",
		},
		Info: "hello world",
	}
	log.RegisterType("info", entry)

	line := string(log.Marshal(entry, nil))
	out := log.Viewable(line)

	if !strings.Contains(out, "hello world") {
		t.Errorf("expected 'hello world' in output, got: %s", out)
	}
	if !strings.Contains(out, "trace-123") {
		t.Errorf("expected 'trace-123' in output, got: %s", out)
	}
}

func TestToJSON(t *testing.T) {
	entry := &log.InfoLog{
		BaseLog: log.BaseLog{
			LogName: "test-app",
			LogType: "info",
			LogTime: 1714896000000000000,
			TraceId: "trace-123",
		},
		Info: "hello world",
	}
	entry.Extra = map[string]any{"key": "value"}
	log.RegisterType("info", entry)

	line := string(log.Marshal(entry, nil))
	jsonStr := log.ToJSON(line)

	if !strings.Contains(jsonStr, `"Info":"hello world"`) {
		t.Errorf("expected Info field in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"TraceId":"trace-123"`) {
		t.Errorf("expected TraceId field in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"key":"value"`) {
		t.Errorf("expected Extra fields merged in JSON, got: %s", jsonStr)
	}
}

func TestLoadMeta(t *testing.T) {
	// Create a temporary meta file
	metaData := `{"test-type":[{"index":0,"name":"Field1"},{"index":1,"name":"Field2"}]}`
	_ = os.WriteFile(".test_meta.json", []byte(metaData), 0644)
	defer os.Remove(".test_meta.json")

	err := log.LoadMeta(".test_meta.json")
	if err != nil {
		t.Fatalf("failed to load meta: %v", err)
	}

	meta := log.GetMeta("test-type")
	if len(meta) != 2 {
		t.Errorf("expected 2 meta fields, got %d", len(meta))
	}
	if meta[0].Name != "Field1" {
		t.Errorf("expected Field1, got %s", meta[0].Name)
	}
}

type EnhancedLog struct {
	log.BaseLog
	App            string            `log:"pos:10,withoutkey:true"`
	Node           string            `log:"pos:11,attachBefore,withoutkey:true"`
	RequestHeaders map[string]string `log:"pos:13,keyname:reqH"`
	ClientIP       string            `log:"pos:12,keyname:ip"`
	Tags           []string          `log:"pos:14"`
}

func (l *EnhancedLog) Reset() {
	l.BaseLog.Reset()
	l.App = ""
	l.Node = ""
	l.ClientIP = ""
	l.RequestHeaders = nil
	l.Tags = nil
}

func TestEnhancedViewable(t *testing.T) {
	entry := &EnhancedLog{
		BaseLog: log.BaseLog{
			LogType: "enhanced",
			LogTime: 1714896000000000000,
		},
		App:      "MyApp",
		Node:     "Node1",
		ClientIP: "127.0.0.1",
		RequestHeaders: map[string]string{
			"User-Agent": "Go-http-cli",
		},
		Tags: []string{"tag1", "tag2"},
	}
	log.RegisterType("enhanced", entry)

	line := string(log.Marshal(entry, nil))
	out := log.Viewable(line)

	// Check attachBefore: MyApp:Node1 (since both are withoutkey)
	if !strings.Contains(out, "MyApp:Node1") {
		t.Errorf("expected MyApp:Node1, got: %s", out)
	}

	// Check pos ordering and keyname: ip:127.0.0.1 should come before reqH
	if !strings.Contains(out, "ip:") || !strings.Contains(out, "127.0.0.1") {
		t.Errorf("expected ip:127.0.0.1, got: %s", out)
	}
	if !strings.Contains(out, "reqH:") || !strings.Contains(out, "User-Agent") || !strings.Contains(out, "Go-http-cli") {
		t.Errorf("expected reqH:[ User-Agent:Go-http-cli ], got: %s", out)
	}

	ipIdx := strings.Index(out, "ip:")
	reqHIdx := strings.Index(out, "reqH:")
	if ipIdx > reqHIdx {
		t.Errorf("expected ip to come before reqH, but ipIdx=%d, reqHIdx=%d", ipIdx, reqHIdx)
	}

	// Check array rendering: Tags:[ tag1 tag2 ]
	if !strings.Contains(out, "Tags:") || !strings.Contains(out, "[ tag1 tag2 ]") {
		t.Errorf("expected Tags:[ tag1 tag2 ], got: %s", out)
	}
}

func TestEnhancedToJSON(t *testing.T) {
	entry := &EnhancedLog{
		BaseLog: log.BaseLog{
			LogType: "enhanced",
			LogTime: 1714896000000000000,
		},
		App:      "MyApp",
		Node:     "Node1",
		ClientIP: "127.0.0.1",
		RequestHeaders: map[string]string{
			"User-Agent": "Go-http-cli",
		},
	}
	log.RegisterType("enhanced", entry)

	line := string(log.Marshal(entry, nil))
	jsonStr := log.ToJSON(line)

	// Check keyname in JSON
	if !strings.Contains(jsonStr, `"ip":"127.0.0.1"`) {
		t.Errorf("expected ip field in JSON, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"reqH":{"User-Agent":"Go-http-cli"}`) {
		t.Errorf("expected reqH field in JSON, got: %s", jsonStr)
	}
}

type CallStackLog struct {
	log.BaseLog
	CallStacks []string `log:"pos:6"`
}

func (l *CallStackLog) Reset() {
	l.BaseLog.Reset()
	l.CallStacks = nil
}

func TestCallStacksViewable(t *testing.T) {
	wd, _ := os.Getwd()
	entry := &CallStackLog{
		BaseLog: log.BaseLog{
			LogType: "test_error",
			LogTime: 1714896000000000000,
		},
		CallStacks: []string{wd + "/main.go:10", "/usr/local/go/src/runtime/panic.go:100"},
	}
	log.RegisterType("test_error", entry)

	line := string(log.Marshal(entry, nil))
	out := log.Viewable(line)

	// Check path truncation (should contain relative "main.go:10")
	if !strings.Contains(out, "main.go:10") {
		t.Errorf("expected relative path main.go:10, got: %s", out)
	}
	// Absolute path should be removed if it matches wd
	if strings.Contains(out, wd) {
		t.Errorf("absolute path should be truncated, but still found: %s", out)
	}
}

type PrecisionLog struct {
	log.BaseLog
	Value float64 `log:"pos:6,precision:2"`
}

func (l *PrecisionLog) Reset() {
	l.BaseLog.Reset()
	l.Value = 0
}

func TestPrecisionViewable(t *testing.T) {
	entry := &PrecisionLog{
		BaseLog: log.BaseLog{
			LogType: "precision",
			LogTime: 1714896000000000000,
		},
		Value: 3.14159,
	}
	log.RegisterType("precision", entry)

	line := string(log.Marshal(entry, nil))
	out := log.Viewable(line)

	if !strings.Contains(out, "3.14") || strings.Contains(out, "3.141") {
		t.Errorf("expected 3.14 (precision 2), got: %s", out)
	}
}
