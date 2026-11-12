package log

import (
	"context"
	"errors"
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

// writerService manages the background writing of log entries.
type writerService struct {
	Running    atomic.Bool
	StopChan   chan struct{}
	LogChannel chan logPayload
	Dropped    atomic.Uint64
	Writers    atomic.Value // []Writer
	WriterLock sync.Mutex

	Files     map[string]*FileWriter
	FilesLock sync.RWMutex
}

var (
	// WriterService is the global instance of defaultService.
	WriterService = newWriterService()
)

func newWriterService() *writerService {
	service := &writerService{
		LogChannel: make(chan logPayload, 10000),
		Files:      make(map[string]*FileWriter),
	}
	service.Writers.Store([]Writer{})
	return service
}

// ConsoleWriter 控制台写入器
type ConsoleWriter struct {
}

func (w *ConsoleWriter) Log(entry LogEntry, data []byte) {
	fmt.Println(Viewable(string(data)))
}

func (w *ConsoleWriter) Run() {
}

func init() {
	RegisterWriterMaker("console", func(conf *Config) Writer {
		return &ConsoleWriter{}
	})
}

// writeAsync 异步写入日志
func writeAsync(payload logPayload) {
	defer func() {
		recover()
	}()
	if !WriterService.Running.Load() {
		return
	}
	select {
	case WriterService.LogChannel <- payload:
	default:
		// 丢弃或处理过载
		dropped := WriterService.Dropped.Add(1)
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
	return WriterService.Dropped.Load()
}

// Start implements starter.Service interface.
func (s *writerService) Start(_ context.Context, _ *Logger) error {
	s.FilesLock.Lock()
	defer s.FilesLock.Unlock()
	if !s.Running.CompareAndSwap(false, true) {
		return nil
	}
	s.LogChannel = make(chan logPayload, 10000)
	s.StopChan = make(chan struct{})
	go s.writerRunner()
	return nil
}

// Stop implements starter.Service interface.
func (s *writerService) Stop(_ context.Context) error {
	if s.Running.CompareAndSwap(true, false) {
		close(s.LogChannel)
		if s.StopChan != nil {
			<-s.StopChan
			s.StopChan = nil
		}
	}
	s.FilesLock.Lock()
	defer s.FilesLock.Unlock()
	var failures []error
	for _, f := range s.Files {
		f.Close()
		if err := f.Err(); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

// Status implements starter.Service interface.
func (s *writerService) Status() (string, error) {
	if s.Running.Load() {
		return fmt.Sprintf("queue: %d, dropped: %d", len(s.LogChannel), s.Dropped.Load()), nil
	}
	return "stopped", nil
}
func (s *writerService) writerRunner() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	defer func() {
		if s.StopChan != nil {
			close(s.StopChan)
		}
	}()

	for {
		select {
		case payload, ok := <-s.LogChannel:
			if !ok {
				s.flushWriters()
				return
			}
			s.processLog(payload)

			// 尝试批量处理更多日志
			batchCount := 0
			for batchCount < 100 {
				select {
				case nextPayload, nextOk := <-s.LogChannel:
					if !nextOk {
						s.flushWriters()
						return
					}
					s.processLog(nextPayload)
					batchCount++
				default:
					batchCount = 100 // break outer loop
				}
			}
		case <-ticker.C:
			s.flushWriters()
		}
	}
}

func (s *writerService) processLog(payload logPayload) {
	// 精准路由：根据包裹信息决定写入目标
	if payload.writer != nil {
		payload.writer.Log(payload.entry, payload.buf)
	} else if payload.file != nil {
		payload.file.Write(time.Now(), payload.buf)
	} else {
		// 默认回退到控制台打印
		fmt.Println(Viewable(string(payload.buf)))
	}
}

func (s *writerService) flushWriters() {
	curWriters, _ := s.Writers.Load().([]Writer)
	for _, w := range curWriters {
		w.Run()
	}

	s.FilesLock.RLock()
	for _, f := range s.Files {
		f.Run()
	}
	s.FilesLock.RUnlock()
}
