package log

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// FileWriter 文件写入器
type FileWriter struct {
	mu          sync.Mutex
	noCompress  bool
	dictionary  *archiveDictionary
	dictFP      *os.File
	pending     [][]byte
	pendingSize int
	lastError   error
	fileName    string
	lastSplit   string
	splitTag    string
	fp          *os.File
	bufWriter   *bufio.Writer
}

// Write 写入缓冲区；与同步写入及 Flush 共用锁。
func (f *FileWriter) Write(tm time.Time, data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.write(tm, data)
}

// writeSync 在异步服务未运行时立即刷新，保证启动和停止后的日志落盘。
func (f *FileWriter) writeSync(tm time.Time, data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.write(tm, data)
	f.flush()
}

func (f *FileWriter) write(tm time.Time, data []byte) {
	nowSplit := tm.Format(f.splitTag)

	// 1. 文件切割逻辑 (按天/按小时流转)
	if f.lastSplit != nowSplit || f.fp == nil {
		f.rotateFile(nowSplit)
	}

	if len(data) > maxArchiveLine {
		f.lastError = fmt.Errorf("log line exceeds 16 MiB")
		fmt.Fprintln(os.Stderr, f.lastError)
		return
	}
	if !f.noCompress {
		if f.fp != nil && f.dictionary != nil {
			f.pending = append(f.pending, append([]byte(nil), data...))
			f.pendingSize += len(data)
			if f.pendingSize >= 64*1024 {
				f.flush()
			}
			return
		}
		fmt.Println(string(data))
		return
	}
	// 2. 直接写内存缓冲区
	if f.bufWriter != nil {
		_, err := f.bufWriter.Write(data)
		if err == nil {
			f.bufWriter.WriteByte('\n') // 追加换行
			return
		}
	}

	// 降级：如果文件句柄异常，打印到控制台避免丢失
	fmt.Println(string(data))
}

// Run 充当 Flush 的角色，由外层的 200ms Ticker 定时调用
func (f *FileWriter) Run() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flush()
}

// 处理文件轮转的内部方法
func (f *FileWriter) flush() {
	if f.bufWriter != nil {
		if err := f.bufWriter.Flush(); err != nil {
			f.lastError = err
			fmt.Fprintln(os.Stderr, "log flush:", err)
		}
	}
	if len(f.pending) > 0 && f.fp != nil {
		if err := writeArchiveBlock(f.fp, f.dictFP, f.dictionary, f.pending); err != nil {
			f.lastError = err
			fmt.Fprintln(os.Stderr, "log archive:", err)
			for _, line := range f.pending {
				fmt.Println(string(line))
			}
			f.closeFiles()
		}
		f.pending = nil
		f.pendingSize = 0
	}
}

// Err returns the most recent output failure, including asynchronous failures.
func (f *FileWriter) Err() error { f.mu.Lock(); defer f.mu.Unlock(); return f.lastError }

func (f *FileWriter) closeFiles() {
	if f.dictFP != nil {
		_ = f.dictFP.Close()
		f.dictFP = nil
	}
	if f.fp != nil {
		_ = f.fp.Close()
		f.fp = nil
	}
	f.bufWriter = nil
	f.dictionary = nil
}

func (f *FileWriter) rotateFile(nowSplit string) {
	f.flush()
	f.closeFiles()
	path := f.fileName
	if f.splitTag != "" {
		path += "." + nowSplit
	}
	if !f.noCompress && !strings.HasSuffix(path, ".gz") {
		path += ".gz"
	}
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err == nil {
		err = lockLogFile(fp)
	}
	if err == nil && !f.noCompress {
		dictionaryName := dictionaryPath(path)
		f.dictionary, err = loadDictionary(dictionaryName, true)
		if os.IsNotExist(err) {
			stat, statErr := fp.Stat()
			if statErr != nil {
				err = statErr
			} else if stat.Size() != 0 {
				err = fmt.Errorf("missing dictionary for %s", path)
			} else {
				f.dictionary, err = createDictionary(dictionaryName)
			}
		}
		if err == nil {
			var end int64
			var blocks []archiveBlock
			blocks, end, err = scanArchive(fp, f.dictionary.Identity)
			if err == nil && len(blocks) > 0 {
				err = readArchiveBlock(fp, blocks[len(blocks)-1], f.dictionary, func([]byte) error { return nil })
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				err = fp.Truncate(end)
			}
		}
		if err == nil {
			f.dictFP, err = os.OpenFile(dictionaryName, os.O_WRONLY|os.O_APPEND, 0600)
		}
	}
	if err != nil {
		if fp != nil {
			_ = fp.Close()
		}
		f.closeFiles()
		f.lastError = err
		fmt.Fprintln(os.Stderr, "failed to open log:", path, err)
		return
	}
	f.fp = fp
	f.lastSplit = nowSplit
	if f.noCompress {
		f.bufWriter = bufio.NewWriterSize(fp, 64*1024)
	}
}

func (f *FileWriter) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flush()
	f.closeFiles()
}
