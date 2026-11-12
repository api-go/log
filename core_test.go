package log

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoggerCore_Initialization(t *testing.T) {
	// Test default configuration initialization
	conf := Config{NoCompress: true}
	logger := NewLogger(conf)

	if logger.config.Level != "info" {
		t.Errorf("Expected default level 'info', got %s", logger.config.Level)
	}
	if logger.config.Name == "" {
		t.Error("Expected default logger name, got empty")
	}
}

func TestLoggerCore_SplitFileInitialization(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "logs", "server.log")
	logger := NewLogger(Config{NoCompress: true,
		File:     logFile,
		SplitTag: "20060102",
	})

	if logger.file == nil {
		t.Fatal("expected split file writer to be initialized")
	}
	if _, err := os.Stat(filepath.Dir(logFile)); err != nil {
		t.Fatalf("expected parent directory to be created: %v", err)
	}
}

func TestLoggerCore_SplitFileWritesBeforeWriterServiceStarts(t *testing.T) {
	if WriterService.Running.Load() {
		t.Skip("writer service already running")
	}

	logFile := filepath.Join(t.TempDir(), "logs", "server.log")
	logger := NewLogger(Config{NoCompress: true,
		File:     logFile,
		SplitTag: "20060102",
	})
	logger.Info("before writer service starts")
	logger.file.Run()

	data, err := os.ReadFile(logFile + "." + time.Now().Format("20060102"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "before writer service starts") {
		t.Fatalf("expected pre-start log to be written to split file, got %q", data)
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	oldConfig := defaultConfig
	oldWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defaultConfig = oldConfig
		_ = os.Chdir(oldWorkingDir)
	})

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "log.yml"), []byte("level: warning\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defaultConfig = Config{NoCompress: true,
		Name:  "built-in",
		Level: "debug",
		File:  "logs/server.log",
	}
	conf := loadDefaultConfig()

	if conf.Name != "built-in" {
		t.Fatalf("expected default name to be retained, got %q", conf.Name)
	}
	if conf.File != "logs/server.log" {
		t.Fatalf("expected default file to be retained, got %q", conf.File)
	}
	if conf.Level != "warning" {
		t.Fatalf("expected file config to override level, got %q", conf.Level)
	}
}

func TestSetDefaultConfig(t *testing.T) {
	oldConfig := defaultConfig
	oldLogger := DefaultLogger
	t.Cleanup(func() {
		defaultConfig = oldConfig
		DefaultLogger = oldLogger
		DefaultLogger.RedirectStdLog()
	})

	SetDefaultConfig(Config{NoCompress: true, Name: "application-default", Level: "debug"})

	if DefaultLogger.config.Name != "application-default" {
		t.Fatalf("expected rebuilt default logger name, got %q", DefaultLogger.config.Name)
	}
	if DefaultLogger.config.Level != "debug" {
		t.Fatalf("expected rebuilt default logger level, got %q", DefaultLogger.config.Level)
	}
}

func TestSetDefaultConfigKeepsLoggerWhenEffectiveConfigIsUnchanged(t *testing.T) {
	oldConfig := defaultConfig
	oldLogger := DefaultLogger
	t.Cleanup(func() {
		defaultConfig = oldConfig
		DefaultLogger = oldLogger
		DefaultLogger.RedirectStdLog()
	})

	defaultConfig = Config{NoCompress: true}
	DefaultLogger = NewLogger(Config{NoCompress: true, Name: "configured", Level: "info"})
	currentLogger := DefaultLogger

	SetDefaultConfig(Config{NoCompress: true, Name: "configured"})

	if DefaultLogger != currentLogger {
		t.Fatal("expected unchanged effective config to keep the existing logger")
	}
}

func TestLoggerCore_Concurrency(t *testing.T) {
	conf := Config{NoCompress: true,
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
