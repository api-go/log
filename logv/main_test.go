package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestProcessAcceptsLargeLogLine(t *testing.T) {
	line := strings.Repeat("x", 256*1024)
	var output bytes.Buffer

	if err := process(strings.NewReader(line+"\n"), &output, false); err != nil {
		t.Fatalf("process returned an error for a large valid line: %v", err)
	}
	if got := output.String(); got != line+"\n" {
		t.Fatalf("unexpected output length: got %d, want %d", len(got), len(line)+1)
	}
}

func TestProcessRejectsLogLineOverLimit(t *testing.T) {
	line := strings.Repeat("x", maxLogLineSize+1)

	err := process(strings.NewReader(line), io.Discard, false)
	if err == nil {
		t.Fatal("process accepted a line over the configured limit")
	}
	if !strings.Contains(err.Error(), "maximum log line size 16 MiB") {
		t.Fatalf("error does not describe the configured limit: %v", err)
	}
}
