package log

import (
	"testing"
)

func TestSetDefaultName(t *testing.T) {
	oldName := getDefaultName()
	defer SetDefaultName(oldName)

	newName := "test-service-name"
	SetDefaultName(newName)

	if getDefaultName() != newName {
		t.Errorf("GetDefaultName() = %v, want %v", getDefaultName(), newName)
	}

	if DefaultLogger.config.Name != newName {
		t.Errorf("DefaultLogger.config.Name = %v, want %v", DefaultLogger.config.Name, newName)
	}
}
