package log_test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/api-go/cast"
	"github.com/api-go/file"
	"github.com/api-go/log"
)

func TestFileRoutingLifecycle(t *testing.T) {
	for _, splitTag := range []string{"", "2006"} {
		t.Run("split="+splitTag, func(t *testing.T) {
			log.WriterService.Stop(nil)
			t.Cleanup(func() { log.WriterService.Stop(nil) })
			path := filepath.Join(t.TempDir(), "logs", "app.log")
			logger := log.NewLogger(log.Config{File: path, SplitTag: splitTag})
			logger.Info("before-start")
			target := path
			if splitTag != "" {
				target += "." + time.Now().Format(splitTag)
			}
			if data, err := file.Read(target); err != nil || !strings.Contains(data, "before-start") {
				t.Fatalf("startup record not immediately readable: %q, %v", data, err)
			}
			if err := log.WriterService.Start(nil, nil); err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			for i := 0; i < 40; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					shared := log.NewLogger(log.Config{File: path, SplitTag: splitTag})
					shared.Info("during-service")
				}()
			}
			wg.Wait()
			if err := log.WriterService.Stop(nil); err != nil {
				t.Fatal(err)
			}
			logger.Info("after-stop")
			// Repeated service starts preserve the same file destination.
			if err := log.WriterService.Start(nil, nil); err != nil {
				t.Fatal(err)
			}
			logger.Info("after-restart")
			if err := log.WriterService.Stop(nil); err != nil {
				t.Fatal(err)
			}
			data, err := file.Read(target)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSpace(data), "\n")
			if len(lines) != 43 {
				t.Fatalf("expected 43 file records, got %d", len(lines))
			}
			for _, line := range lines {
				var arr []any
				if err := cast.UnmarshalJSON([]byte(line), &arr); err != nil {
					t.Fatalf("expected pure JSON array: %q: %v", line, err)
				}
				if log.Viewable(line) == line {
					t.Fatalf("viewer did not render %q", line)
				}
				if !strings.HasPrefix(log.ToJSON(line), "{") {
					t.Fatalf("JSON viewer did not convert %q", line)
				}
			}
			if file.Exists(path + ".") {
				t.Fatal("empty split tag must not append a dot")
			}
			if !strings.Contains(lines[0], "before-start") || !strings.Contains(lines[41], "after-stop") || !strings.Contains(lines[42], "after-restart") {
				t.Fatal("lifecycle records missing or out of order")
			}
		})
	}
}
