package log_test

import (
	"bytes"
	"context"
	"github.com/api-go/log"
	"testing"
)

func TestBareTailFollowJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	previous := log.DefaultConfig()
 log.SetDefaultConfig(log.Config{File: t.TempDir()+"/test.log"})
 defer log.SetDefaultConfig(previous)
 var output bytes.Buffer
	err := log.RunCommand(ctx, "log", []string{"-tail", "-f", "-json"}, bytes.NewReader(nil), &output)
	if err != context.Canceled {
		t.Fatalf("expected canceled follow, got %v", err)
	}
	if bytes.Contains(output.Bytes(), []byte("\x1b[")) {
		t.Fatal("JSON contains ANSI")
	}
}
