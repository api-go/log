package log

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestLoggerReliability(t *testing.T) {
	logFile := "reliability.log"
	defer os.Remove(logFile)

	logger := NewLogger(Config{NoCompress: true,
		File: logFile,
	})

	const count = 1000
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(idx int) {
			defer wg.Done()
			logger.Info("reliability", "index", idx)
		}(i)
	}

	wg.Wait()
	WriterService.Stop(nil)

	file, err := os.Open(logFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	found := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "reliability") {
			lines++
			found[line] = true
		}
	}

	if lines != count {
		t.Errorf("Expected %d log lines, got %d", count, lines)
	}
}
