package log

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

// FileWriter 文件写入器
type FileWriter struct {
	fileName  string
	lastSplit string
	splitTag  string
	fp        *os.File
	bufWriter *bufio.Writer
}

var (
	files     = make(map[string]*FileWriter)
	filesLock sync.RWMutex
)

// Write 由外层的 writerRunner 单协程调用，绝对并发安全，无需加锁
func (f *FileWriter) Write(tm time.Time, data []byte) {
	nowSplit := tm.Format(f.splitTag)
	
	// 1. 文件切割逻辑 (按天/按小时流转)
	if f.lastSplit != nowSplit || f.fp == nil {
		f.rotateFile(nowSplit)
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
	if f.bufWriter != nil {
		_ = f.bufWriter.Flush()
	}
}

// 处理文件轮转的内部方法
func (f *FileWriter) rotateFile(nowSplit string) {
	if f.bufWriter != nil {
		_ = f.bufWriter.Flush()
	}
	if f.fp != nil {
		_ = f.fp.Close()
	}

	var err error
	f.fp, err = os.OpenFile(f.fileName+"."+nowSplit, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		// 分配 64KB 缓冲区
		f.bufWriter = bufio.NewWriterSize(f.fp, 64*1024)
		f.lastSplit = nowSplit
	} else {
		f.bufWriter = nil
		fmt.Printf("failed to open log file: %s.%s, error: %v\n", f.fileName, nowSplit, err)
	}
}

func (f *FileWriter) Close() {
	if f.bufWriter != nil {
		_ = f.bufWriter.Flush()
		f.bufWriter = nil
	}
	if f.fp != nil {
		_ = f.fp.Close()
		f.fp = nil
	}
}
