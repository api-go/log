package log

import (
	"sync"
	"testing"
)

func TestLoggerCore_Initialization(t *testing.T) {
	// Test default configuration initialization
	conf := Config{}
	logger := NewLogger(conf)

	if logger.config.Level != "info" {
		t.Errorf("Expected default level 'info', got %s", logger.config.Level)
	}
	if logger.config.Name == "" {
		t.Error("Expected default logger name, got empty")
	}
}

func TestLoggerCore_Concurrency(t *testing.T) {
	conf := Config{
		Name:  "concurrent-test",
		Level: "info",
	}
	logger := NewLogger(conf)

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Info("concurrent message", "id", id)
		}(i)
	}

	wg.Wait()
}
