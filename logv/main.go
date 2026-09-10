package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"apigo.cc/go/log"
)

const maxLogLineSize = 16 * 1024 * 1024

func main() {
	os.Exit(run())
}

func run() int {
	jsonMode := flag.Bool("json", false, "output in JSON format")
	helpMode := flag.Bool("h", false, "show help")
	flag.BoolVar(helpMode, "help", false, "show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: logv [options] [file1 file2 ...]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  1. 彩色可视化实时日志:\n")
		fmt.Fprintf(os.Stderr, "     tail -f app.log | logv\n")
		fmt.Fprintf(os.Stderr, "  2. 格式化查看历史日志文件:\n")
		fmt.Fprintf(os.Stderr, "     logv error.log\n")
		fmt.Fprintf(os.Stderr, "  3. 还原为标准 JSON 格式 (供 Filebeat / Logstash 收集):\n")
		fmt.Fprintf(os.Stderr, "     tail -f app.log | logv -json\n")
		fmt.Fprintf(os.Stderr, "  4. 批量转换日志文件为 JSON:\n")
		fmt.Fprintf(os.Stderr, "     logv -json app.log.2026* > all_logs.json\n")
	}

	flag.Parse()

	if *helpMode {
		flag.Usage()
		return 0
	}

	// Try to load meta file from current directory
	_ = log.LoadMeta(".log.meta.json")

	args := flag.Args()
	if len(args) == 0 {
		// Check if stdin is a terminal
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// Stdin is a terminal and no files provided, show usage
			flag.Usage()
			return 0
		}
		if err := process(os.Stdin, os.Stdout, *jsonMode); err != nil {
			fmt.Fprintf(os.Stderr, "logv: %v\n", err)
			return 1
		}
	} else {
		exitCode := 0
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "logv: %v\n", err)
				exitCode = 1
				continue
			}
			err = process(f, os.Stdout, *jsonMode)
			closeErr := f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "logv: %s: %v\n", arg, err)
				exitCode = 1
			} else if closeErr != nil {
				fmt.Fprintf(os.Stderr, "logv: %s: %v\n", arg, closeErr)
				exitCode = 1
			}
		}
		return exitCode
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
