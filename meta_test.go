package log

import (
	"encoding/json"
	"os"
	"testing"
)

type MockBaseLog struct {
	BaseField1 string `log:"pos:1,color:red"`
	BaseField2 int    `log:"pos:2,withoutkey:true"`
}

type MockInfoLog struct {
	MockBaseLog
	Message string `log:"pos:3"`
	Extra   map[string]any
}

type MockErrorLog struct {
	MockBaseLog
	Error      string `log:"pos:3,color:red"`
	CallStacks []string
	Extra      map[string]any
}

func TestMetaExtraction(t *testing.T) {
	// Setup custom meta file path for testing
	SetMetaFilePath(".test.meta.json")
	defer os.Remove(".test.meta.json")

	RegisterType("mock_info", MockInfoLog{})
	RegisterType("mock_error", MockErrorLog{})

	infoMeta := GetMeta("mock_info")
	if len(infoMeta) != 4 { // BaseField1, BaseField2, Message, Extra
		t.Fatalf("expected 4 fields for mock_info, got %d", len(infoMeta))
	}

	if infoMeta[0].Name != "BaseField1" || infoMeta[0].Color != "red" {
		t.Errorf("unexpected meta for BaseField1: %+v", infoMeta[0])
	}
	if infoMeta[1].Name != "BaseField2" || infoMeta[1].WithoutKey != true {
		t.Errorf("unexpected meta for BaseField2: %+v", infoMeta[1])
	}
	if infoMeta[2].Name != "Message" {
		t.Errorf("unexpected meta for Message: %+v", infoMeta[2])
	}
	if infoMeta[3].Name != "Extra" {
		t.Errorf("unexpected meta for Extra: %+v", infoMeta[3])
	}

	errorMeta := GetMeta("mock_error")
	if len(errorMeta) != 5 { // BaseField1, BaseField2, Error, CallStacks, Extra
		t.Fatalf("expected 5 fields for mock_error, got %d", len(errorMeta))
	}
	
	if errorMeta[2].Name != "Error" || errorMeta[2].Color != "red" {
		t.Errorf("unexpected meta for Error: %+v", errorMeta[2])
	}
	if errorMeta[3].Name != "CallStacks" {
		t.Errorf("unexpected meta for CallStacks: %+v", errorMeta[3])
	}
	if errorMeta[4].Name != "Extra" {
		t.Errorf("unexpected meta for Extra: %+v", errorMeta[4])
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
