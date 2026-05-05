package main

import (
	"bufio"
	"fmt"
	"os"

	"apigo.cc/go/log"
)

func main() {
	// Ensure built-in types are registered to get basic meta if .log.meta.json is missing
	// log package init() handles most of it, but we can also just run it.

	// Reading from stdin
	scanner := bufio.NewScanner(os.Stdin)

	// Optional: Adjust max token size if log lines are extremely long
	// buf := make([]byte, 0, 64*1024)
	// scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		
		// Render and print the log line
		rendered := log.Viewable(line)
		fmt.Println(rendered)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "logv: error reading standard input: %v\n", err)
		os.Exit(1)
	}
}
