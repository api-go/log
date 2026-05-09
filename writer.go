package log

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Writer 日志写入接口
type Writer interface {
	Log(LogEntry, []byte)
	Run()
}

// logPayload 包含路由信息的包裹
type logPayload struct {
	entry  LogEntry
	buf    []byte
	writer Writer      // 目标自定义 Writer
	file   *FileWriter // 目标文件 Writer
}

var (
	writerRunning  atomic.Bool
	writerLock     sync.Mutex // 仅用于注册时锁定
	writerStopChan chan bool
	writers        atomic.Value // 存储 []Writer
	logChannel     chan logPayload
	droppedLogs    atomic.Uint64
)

// ConsoleWriter 控制台写入器
type ConsoleWriter struct {
}

func (w *ConsoleWriter) Log(entry LogEntry, data []byte) {
	fmt.Println(Viewable(string(data)))
}

func (w *ConsoleWriter) Run() {
}

func init() {
	logChannel = make(chan logPayload, 10000)
	writers.Store([]Writer{})
	RegisterWriterMaker("console", func(conf *Config) Writer {
		return &ConsoleWriter{}
	})
}

// WriteAsync 异步写入日志
func WriteAsync(payload logPayload) {
	defer func() {
		recover()
	}()
	if !writerRunning.Load() {
		return
	}
	select {
	case logChannel <- payload:
	default:
		// 丢弃或处理过载
		dropped := droppedLogs.Add(1)
		if dropped%1000 == 1 {
			if DefaultLogger != nil {
				// 注意：这里可能会产生递归调用，但 select default 保证了不会死锁
				DefaultLogger.Error(fmt.Sprintf("log channel full, dropped %d logs", dropped))
			}
		}
	}
}

// GetDroppedLogs 获取被丢弃的日志数量
func GetDroppedLogs() uint64 {
	return droppedLogs.Load()
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
		case payload, ok := <-logChannel:
			if !ok {
				flushWriters()
				return
			}
			processLog(payload)

			// 尝试批量处理更多日志
			batchCount := 0
			for batchCount < 100 {
				select {
				case nextPayload, nextOk := <-logChannel:
					if !nextOk {
						flushWriters()
						return
					}
					processLog(nextPayload)
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

func processLog(payload logPayload) {
	// 精准路由：根据包裹信息决定写入目标
	if payload.writer != nil {
		payload.writer.Log(payload.entry, payload.buf)
	} else if payload.file != nil {
		payload.file.Write(time.Now(), payload.buf)
	}
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
