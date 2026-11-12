package log_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/api-go/file"
	"github.com/api-go/log"
)

func archiveFixture(t *testing.T, n int) string {
	t.Helper()
	_ = log.WriterService.Stop(nil)
	path := filepath.Join(t.TempDir(), "server.log")
	logger := log.NewLogger(log.Config{File: path})
	for i := 0; i < n; i++ {
		logger.Info(fmt.Sprintf("message-%04d", i), "ua", strings.Repeat("browser-agent/", 30), "nested", map[string]any{"password": "must-not-leak", "escaped": "quotes \" and \\ and 中文"})
	}
	if err := log.WriterService.Stop(nil); err != nil {
		t.Fatal(err)
	}
	return path + ".gz"
}

func TestArchiveRoundTripAndRanges(t *testing.T) {
	path := archiveFixture(t, 130)
	for _, tc := range []struct {
		opts        log.QueryOptions
		count       int
		first, last string
	}{
		{log.QueryOptions{Raw: true}, 130, "message-0000", "message-0129"},
		{log.QueryOptions{Raw: true, Tail: 3}, 3, "message-0127", "message-0129"},
		{log.QueryOptions{JSON: true, Head: 2}, 2, "message-0000", "message-0001"},
	} {
		var out bytes.Buffer
		if err := log.QueryFile(context.Background(), path, &out, tc.opts); err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(lines) != tc.count || !strings.Contains(lines[0], tc.first) || !strings.Contains(lines[len(lines)-1], tc.last) {
			t.Fatalf("unexpected range: %s", out.String())
		}
		if strings.Contains(out.String(), "must-not-leak") {
			t.Fatal("sensitive field leaked")
		}
	}
	// Reopen appends using the existing pool and restores previous references.
	logger := log.NewLogger(log.Config{File: strings.TrimSuffix(path, ".gz")})
	logger.Info("after-reopen")
	if err := log.WriterService.Stop(nil); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := log.QueryFile(context.Background(), path, &out, log.QueryOptions{Tail: 1, Raw: true}); err != nil || !strings.Contains(out.String(), "after-reopen") {
		t.Fatalf("reopen: %s %v", out.String(), err)
	}
}

func TestArchiveRejectsMissingMismatchedAndCorruptDictionary(t *testing.T) {
	path := archiveFixture(t, 3)
	dict := strings.TrimSuffix(path, ".gz") + ".dict"
	saved, err := file.ReadBytes(dict)
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Remove(dict); err != nil {
		t.Fatal(err)
	}
	if err = log.QueryFile(context.Background(), path, &bytes.Buffer{}, log.QueryOptions{}); err == nil {
		t.Fatal("missing dictionary accepted")
	}
	changed := bytes.Clone(saved)
	changed[8] ^= 1
	if err = file.WriteBytes(dict, changed); err != nil {
		t.Fatal(err)
	}
	if err = log.QueryFile(context.Background(), path, &bytes.Buffer{}, log.QueryOptions{}); err == nil {
		t.Fatal("mismatched dictionary accepted")
	}
	changed = bytes.Clone(saved)
	changed[28] ^= 1
	if err = file.WriteBytes(dict, changed); err != nil {
		t.Fatal(err)
	}
	if err = log.QueryFile(context.Background(), path, &bytes.Buffer{}, log.QueryOptions{}); err == nil {
		t.Fatal("corrupt dictionary accepted")
	}
}

func TestArchiveIncompleteTailRecovery(t *testing.T) {
	path := archiveFixture(t, 3)
	fp, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fp.Write([]byte{31, 139, 8})
	fp.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err = log.QueryFile(context.Background(), path, &bytes.Buffer{}, log.QueryOptions{}); err == nil {
		t.Fatal("truncated file accepted")
	}
	logger := log.NewLogger(log.Config{File: strings.TrimSuffix(path, ".gz")})
	logger.Info("recovered")
	if err = log.WriterService.Stop(nil); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err = log.QueryFile(context.Background(), path, &out, log.QueryOptions{Tail: 1, Raw: true}); err != nil || !strings.Contains(out.String(), "recovered") {
		t.Fatalf("recovery %s %v", out.String(), err)
	}
}

func TestCommandRangesAndCancellation(t *testing.T) {
	path := archiveFixture(t, 4)
	var out bytes.Buffer
	if err := log.RunCommand(context.Background(), "log", []string{path, "-tail", "2"}, nil, &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "message-0001") || !strings.Contains(out.String(), "message-0002") {
		t.Fatal(out.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := log.QueryFile(ctx, path, &out, log.QueryOptions{Follow: true}); err != context.Canceled {
		t.Fatal(err)
	}
}

func BenchmarkArchiveWrite(b *testing.B) {
	_ = log.WriterService.Stop(nil)
	logger := log.NewLogger(log.Config{File: filepath.Join(b.TempDir(), "bench.log")})
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		logger.Info("proxy completed", "ua", "Mozilla/5.0 repeated browser agent", "status", 200)
	}
	b.StopTimer()
	if err := log.WriterService.Stop(nil); err != nil {
		b.Fatal(err)
	}
}

type followOutput struct {
	bytes.Buffer
	logger   *log.Logger
	cancel   context.CancelFunc
	appended bool
}

func (w *followOutput) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	if !w.appended {
		w.appended = true
		w.logger.Info("follow-appended")
	} else if bytes.Contains(p, []byte("follow-appended")) {
		w.cancel()
	}
	return n, err
}

func TestArchiveFollowAndCompressedStdin(t *testing.T) {
	path := archiveFixture(t, 3)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out := &followOutput{logger: log.NewLogger(log.Config{File: strings.TrimSuffix(path, ".gz")}), cancel: cancel}
	err := log.QueryFile(ctx, path, out, log.QueryOptions{Tail: 1, Follow: true, Raw: true})
	if !errors.Is(err, context.Canceled) || !strings.Contains(out.String(), "follow-appended") {
		t.Fatalf("follow: %v %s", err, out.String())
	}
	if err = log.WriterService.Stop(nil); err != nil {
		t.Fatal(err)
	}
	compressed, err := file.ReadBytes(path)
	if err != nil {
		t.Fatal(err)
	}
	var restored bytes.Buffer
	if err = log.FormatArchiveStream(context.Background(), bytes.NewReader(compressed), &restored, strings.TrimSuffix(path, ".gz")+".dict", log.QueryOptions{Raw: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(restored.String(), "follow-appended") {
		t.Fatal("stdin lost appended line")
	}
}
