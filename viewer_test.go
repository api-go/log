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

	line := string(log.ToArrayBytes(entry, nil))
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

	line := string(log.ToArrayBytes(entry, nil))
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
