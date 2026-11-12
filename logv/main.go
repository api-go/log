package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/api-go/log"
)

const maxLogLineSize = 16 * 1024 * 1024

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := log.RunCommand(ctx, "logv", os.Args[1:], os.Stdin, os.Stdout); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "logv:", err)
		return 1
	}
	return 0
}

func process(r io.Reader, w io.Writer, jsonMode bool) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxLogLineSize)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if jsonMode {
			fmt.Fprintln(w, log.ToJSON(line))
		} else {
			fmt.Fprintln(w, log.Viewable(line))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input (maximum log line size %d MiB): %w", maxLogLineSize/(1024*1024), err)
	}
	return nil
}
