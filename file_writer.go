package log

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

type FileLogEntry struct {
	time    time.Time
	message string
}

// FileWriter 文件写入器
type FileWriter struct {
	fileName  string
	lastSplit string
	splitTag  string
	fp        *os.File
	bufWriter *bufio.Writer
	entries   []FileLogEntry
	lock      sync.Mutex
}

var (
	files     = make(map[string]*FileWriter)
	filesLock sync.RWMutex
)

func (f *FileWriter) Write(tm time.Time, str string) {
	f.lock.Lock()
	f.entries = append(f.entries, FileLogEntry{
		time:    tm,
		message: str,
	})
	f.lock.Unlock()
}

func (f *FileWriter) Run() {
	f.lock.Lock()
	var runEntries []FileLogEntry
	if len(f.entries) > 0 {
		runEntries = f.entries
		f.entries = nil
	}
	f.lock.Unlock()

	if len(runEntries) > 0 {
		for _, l := range runEntries {
			nowSplit := l.time.Format(f.splitTag)
			if f.lastSplit != nowSplit || f.fp == nil {
				f.lastSplit = nowSplit
				f.lock.Lock()
				if f.bufWriter != nil {
					_ = f.bufWriter.Flush()
				}
				if f.fp != nil {
					_ = f.fp.Close()
				}
				var err error
				f.fp, err = os.OpenFile(f.fileName+"."+nowSplit, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err == nil {
					f.bufWriter = bufio.NewWriterSize(f.fp, 64*1024)
				} else {
					f.bufWriter = nil
				}
				f.lock.Unlock()
				if err != nil {
					fmt.Printf("failed to open log file: %s.%s, error: %v\n", f.fileName, nowSplit, err)
					continue
				}
			}

			logStr := l.time.Format("2006/01/02 15:04:05.000000") + " " + l.message + "\n"
			f.lock.Lock()
			if f.bufWriter != nil {
				_, err := f.bufWriter.WriteString(logStr)
				if err != nil {
					fmt.Print(logStr)
				}
			} else {
				fmt.Print(logStr)
			}
			f.lock.Unlock()
		}
		f.lock.Lock()
		if f.bufWriter != nil {
			_ = f.bufWriter.Flush()
		}
		f.lock.Unlock()
	}
}

func (f *FileWriter) Close() {
	f.lock.Lock()
	if f.bufWriter != nil {
		_ = f.bufWriter.Flush()
		f.bufWriter = nil
	}
	if f.fp != nil {
		_ = f.fp.Close()
		f.fp = nil
	}
	f.lock.Unlock()
}
