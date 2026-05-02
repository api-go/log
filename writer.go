package log

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Writer 日志写入接口
type Writer interface {
	Log([]byte)
	Run()
}

var (
	writerRunning  atomic.Bool
	writerLock     sync.Mutex // 仅用于注册时锁定
	writerStopChan chan bool
	writers        atomic.Value // 存储 []Writer
	logChannel     chan []byte
)

// ConsoleWriter 控制台写入器
type ConsoleWriter struct {
}

func (w *ConsoleWriter) Log(data []byte) {
	fmt.Println(Viewable(string(data)))
}

func (w *ConsoleWriter) Run() {
}

func init() {
	logChannel = make(chan []byte, 10000)
	writers.Store([]Writer{})
	RegisterWriterMaker("console", func(conf *Config) Writer {
		return &ConsoleWriter{}
	})
}

// WriteAsync 异步写入日志
func WriteAsync(buf []byte) {
	defer func() {
		recover()
	}()
	if !writerRunning.Load() {
		return
	}
	select {
	case logChannel <- buf:
	default:
		// 丢弃或处理过载，此处简单丢弃
	}
}

// Start 启动写入器
func Start() {
	if !writerRunning.CompareAndSwap(false, true) {
		return
	}
	writerStopChan = make(chan bool)
	go writerRunner()
}

// Stop 停止写入器
func Stop() {
	if writerRunning.CompareAndSwap(true, false) {
		close(logChannel)
	}
}

// Wait 等待写入器停止
func Wait() {
	if writerStopChan != nil {
		<-writerStopChan
		writerStopChan = nil
	}
}

func writerRunner() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	defer func() {
		if writerStopChan != nil {
			close(writerStopChan)
		}
	}()

	for {
		select {
		case buf, ok := <-logChannel:
			if !ok {
				flushWriters()
				return
			}
			processLog(buf)

			// 尝试批量处理更多日志
			batchCount := 0
			for batchCount < 100 {
				select {
				case nextBuf, nextOk := <-logChannel:
					if !nextOk {
						flushWriters()
						return
					}
					processLog(nextBuf)
					batchCount++
				default:
					batchCount = 100 // break outer loop
				}
			}
		case <-ticker.C:
			flushWriters()
		}
	}
}

func processLog(buf []byte) {
	// 使用原子读取的 writer 列表
	curWriters, _ := writers.Load().([]Writer)
	for _, w := range curWriters {
		w.Log(buf)
	}

	// 文件写入处理
	filesLock.RLock()
	for _, f := range files {
		f.Write(time.Now(), string(buf))
	}
	filesLock.RUnlock()
}

func flushWriters() {
	curWriters, _ := writers.Load().([]Writer)
	for _, w := range curWriters {
		w.Run()
	}

	filesLock.RLock()
	for _, f := range files {
		f.Run()
	}
	filesLock.RUnlock()
}
