package log

import (
	"encoding/json"
	"testing"
)

type SerializerMockBaseLog struct {
	LogName string `log:"pos:0"`
	LogType string `log:"pos:1"`
	LogTime int64  `log:"pos:2"`
	TraceId string `log:"pos:3"`
}

func (b *SerializerMockBaseLog) IsLogEntry() bool {
	return true
}

func (b *SerializerMockBaseLog) GetBaseLog() *BaseLog {
	return &BaseLog{LogType: b.LogType}
}

func (b *SerializerMockBaseLog) Reset() {
	b.LogName = ""
	b.LogType = ""
	b.LogTime = 0
	b.TraceId = ""
}

type SerializerMockInfoLog struct {
	SerializerMockBaseLog
	Message string `log:"pos:4"`
	Extra   map[string]any `log:"pos:1000"`
}

func (l *SerializerMockInfoLog) Reset() {
	l.SerializerMockBaseLog.Reset()
	l.Message = ""
	if l.Extra == nil {
		l.Extra = make(map[string]any, 8)
	} else {
		clear(l.Extra)
	}
}

func TestToArrayBytes(t *testing.T) {
	entry := &SerializerMockInfoLog{
		SerializerMockBaseLog: SerializerMockBaseLog{
			LogName: "test-app",
			LogType: "mock_info_test",
			LogTime: 1620000000,
			TraceId: "abc-123",
		},
		Message: "Hello, World!",
		Extra: map[string]any{
			"user_id": 42,
		},
	}

	RegisterType("mock_info_test", entry) // trigger meta generation
	
	bytes := ToArrayBytes(entry, nil)
	str := string(bytes)
	t.Logf("Raw log: %s", str)
	
	var arr []any
	err := json.Unmarshal(bytes, &arr)
	if err != nil {
		t.Fatalf("failed to unmarshal generated array: %v, raw: %s", err, str)
	}

	// Indices: 0, 1, 2, 3, 4, 1000(mapped to 5). Total size 6.
	if len(arr) != 6 {
		t.Fatalf("expected 6 elements, got %d. raw: %s", len(arr), str)
	}

	if arr[0] != "test-app" {
		t.Errorf("expected arr[0] == 'test-app', got %v", arr[0])
	}
	if arr[1] != "mock_info_test" {
		t.Errorf("expected arr[1] == 'mock_info_test', got %v", arr[1])
	}
	// JSON numbers are parsed as float64
	if arr[2] != float64(1620000000) {
		t.Errorf("expected arr[2] == 1620000000, got %v", arr[2])
	}
	if arr[3] != "abc-123" {
		t.Errorf("expected arr[3] == 'abc-123', got %v", arr[3])
	}
	if arr[4] != "Hello, World!" {
		t.Errorf("expected arr[4] == 'Hello, World!', got %v", arr[4])
	}
	
	extraMap, ok := arr[5].(map[string]any)
	if !ok {
		t.Fatalf("expected arr[5] to be map[string]any, got %T (value: %v)", arr[5], arr[5])
	}
	if extraMap["user_id"] != float64(42) {
		t.Errorf("expected extraMap['user_id'] == 42, got %v", extraMap["user_id"])
	}
}

func TestToArrayBytes_Desensitize(t *testing.T) {
	entry := &SerializerMockInfoLog{
		SerializerMockBaseLog: SerializerMockBaseLog{
			LogType: "mock_info_test2",
		},
		Message: "Sensitive Info",
		Extra: map[string]any{
			"password": "my-secret-password",
		},
	}
	
	RegisterType("mock_info_test2", entry)
	
	bytes := ToArrayBytes(entry, []string{"password"})
	str := string(bytes)
	
	var arr []any
	err := json.Unmarshal(bytes, &arr)
	if err != nil {
		t.Fatalf("failed to unmarshal generated array: %v, raw: %s", err, str)
	}

	extraMap := arr[5].(map[string]any)
	if extraMap["password"] != "***" {
		t.Errorf("expected password to be desensitized, got %v", extraMap["password"])
	}
}
