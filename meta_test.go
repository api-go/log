package log

import (
	"encoding/json"
	"os"
	"testing"
)

type MockBaseLog struct {
	BaseField1 string `log:"pos:0,color:red"`
	BaseField2 int    `log:"pos:1,withoutkey:true"`
}

func (b *MockBaseLog) Reset() {
	b.BaseField1 = ""
	b.BaseField2 = 0
}

func (b *MockBaseLog) IsLogEntry() bool { return true }
func (b *MockBaseLog) GetBaseLog() *BaseLog { return &BaseLog{} }

type MockInfoLog struct {
	MockBaseLog
	Message string `log:"pos:2"`
	Extra   map[string]any `log:"pos:1000"`
}

func (l *MockInfoLog) Reset() {
	l.MockBaseLog.Reset()
	l.Message = ""
	clear(l.Extra)
}

type MockErrorLog struct {
	MockBaseLog
	Error      string `log:"pos:2,color:red"`
	CallStacks []string `log:"pos:1001"`
	Extra      map[string]any `log:"pos:1000"`
}

func (l *MockErrorLog) Reset() {
	l.MockBaseLog.Reset()
	l.Error = ""
	l.CallStacks = l.CallStacks[:0]
	clear(l.Extra)
}

func TestMetaExtraction(t *testing.T) {
	// Setup custom meta file path for testing
	SetMetaFilePath(".test.meta.json")
	defer os.Remove(".test.meta.json")

	RegisterType("mock_info", MockInfoLog{})
	RegisterType("mock_error", MockErrorLog{})

	infoMeta := GetMeta("mock_info")
	// Index 0, 1, 2 are used, Extra gets max(2)+1=3. Total size 4.
	if len(infoMeta) != 4 {
		t.Fatalf("expected 4 fields for mock_info, got %d", len(infoMeta))
	}

	if infoMeta[0].Name != "BaseField1" || infoMeta[0].Color != "red" {
		t.Errorf("unexpected meta for BaseField1 at index 0: %+v", infoMeta[0])
	}
	if infoMeta[1].Name != "BaseField2" || infoMeta[1].WithoutKey != true {
		t.Errorf("unexpected meta for BaseField2 at index 1: %+v", infoMeta[1])
	}
	if infoMeta[2].Name != "Message" {
		t.Errorf("unexpected meta for Message at index 2: %+v", infoMeta[2])
	}
	if infoMeta[3].Name != "Extra" {
		t.Errorf("unexpected meta for Extra at index 3: %+v", infoMeta[3])
	}

	errorMeta := GetMeta("mock_error")
	// Indices: 0, 1, 2, Extra(3), CallStacks(4). Total size 5.
	if len(errorMeta) != 5 {
		t.Fatalf("expected 5 fields for mock_error, got %d", len(errorMeta))
	}

	if errorMeta[2].Name != "Error" || errorMeta[2].Color != "red" {
		t.Errorf("unexpected meta for Error at index 2: %+v", errorMeta[2])
	}
	if errorMeta[3].Name != "Extra" {
		t.Errorf("unexpected meta for Extra at index 3: %+v", errorMeta[3])
	}
	if errorMeta[4].Name != "CallStacks" {
		t.Errorf("unexpected meta for CallStacks at index 4: %+v", errorMeta[4])
	}

	// Verify file was created and contains correct data
	data, err := os.ReadFile(".test.meta.json")
	if err != nil {
		t.Fatalf("failed to read test meta file: %v", err)
	}

	var registry map[string][]MetaField
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatalf("failed to unmarshal test meta file: %v", err)
	}

	if len(registry) < 2 {
		t.Errorf("expected at least 2 types in registry, got %d", len(registry))
	}
}
