package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"apigo.cc/go/cast"
	"apigo.cc/go/id"
)

var (
	stdLogFilters     []func(string) bool
	stdLogFiltersLock sync.RWMutex
)

// AddStdLogFilter 注册一个标准库日志过滤器。
// 如果返回 true，则该条日志将被静默丢弃。
func AddStdLogFilter(f func(string) bool) {
	stdLogFiltersLock.Lock()
	defer stdLogFiltersLock.Unlock()
	stdLogFilters = append(stdLogFilters, f)
}

// logRedirectWriter 捕获标准 log 包的输出并转发到 structured logger
type logRedirectWriter struct {
	logger *Logger
}

func (w *logRedirectWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}

	stdLogFiltersLock.RLock()
	filters := stdLogFilters
	stdLogFiltersLock.RUnlock()

	for _, f := range filters {
		if f(msg) {
			return len(p), nil
		}
	}

	w.logger.Info(msg)
	return len(p), nil
}

// RedirectStdLog 将标准 log 包的输出重定向到当前 Logger。
// 注意：这会全局修改标准 log 包的配置。
func (logger *Logger) RedirectStdLog() {
	log.SetOutput(&logRedirectWriter{logger: logger})
	log.SetFlags(0) // 禁用标准 log 的时间前缀，由我们的 logger 统一处理
}

// SetStdLogOutput 仅设置输出目标而不修改 Flags
func SetStdLogOutput(w io.Writer) {
	log.SetOutput(w)
}

type Logger struct {
	config        Config
	level         LevelType
	goLogger      *log.Logger
	file          *FileWriter
	writer        Writer
	truncations   []string
	sensitive     map[string]bool
	sensitiveKeys []string
	traceId       string
}

var (
	writerMakers = make(map[string]func(*Config) Writer)
)

func RegisterWriterMaker(name string, f func(*Config) Writer) {
	writerMakers[name] = f
}

func NewLogger(conf Config) *Logger {
	if conf.Level == "" {
		conf.Level = "info"
	}
	if conf.Truncations == "" {
		conf.Truncations = "github.com/, golang.org/, /apigo.cc/"
	}
	if conf.Sensitive == "" {
		conf.Sensitive = LogDefaultSensitive
	}

	if conf.Name == "" {
		conf.Name = getDefaultName()
	}

	logger := Logger{
		truncations: cast.Split(conf.Truncations, ","),
		traceId:     id.Get10Bytes14MPerSecond(),
	}

	if len(conf.Sensitive) > 0 {
		logger.sensitive = make(map[string]bool)
		ss := cast.Split(conf.Sensitive, ",")
		for _, v := range ss {
			f := fixField(v)
			logger.sensitive[f] = true
			logger.sensitiveKeys = append(logger.sensitiveKeys, f)
		}
	}

	switch strings.ToLower(conf.Level) {
	case "debug":
		logger.level = DEBUG
	case "warning":
		logger.level = WARNING
	case "error":
		logger.level = ERROR
	default:
		logger.level = INFO
	}

	if conf.File != "" && conf.File != "console" {
		if strings.Contains(conf.File, "://") {
			writerName := strings.SplitN(conf.File, "://", 2)[0]
			if m, ok := writerMakers[writerName]; ok {
				if w := m(&conf); w != nil {
					logger.writer = w
					WriterService.WriterLock.Lock()
					cur := WriterService.Writers.Load().([]Writer)
					newW := append(cur, w)
					WriterService.Writers.Store(newW)
					WriterService.WriterLock.Unlock()
				}
			}
		} else {
			if conf.SplitTag != "" {
				WriterService.FilesLock.RLock()
				logger.file = WriterService.Files[conf.File+conf.SplitTag]
				WriterService.FilesLock.RUnlock()
				if logger.file == nil {
					logger.file = &FileWriter{
						fileName: conf.File,
						splitTag: conf.SplitTag,
					}
					WriterService.FilesLock.Lock()
					WriterService.Files[conf.File+conf.SplitTag] = logger.file
					WriterService.FilesLock.Unlock()
				}
			} else {
				fp, err := os.OpenFile(conf.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err == nil {
					logger.goLogger = log.New(fp, "", log.Ldate|log.Lmicroseconds)
				}
			}
		}
	}
	logger.config = conf
	return &logger
}

func (logger *Logger) Log(entry LogEntry) {
	logger.asyncWrite(entry)
}

func (logger *Logger) asyncWrite(entry LogEntry) {
	buf := Marshal(entry, logger.sensitiveKeys)
	logger.writeBuf(entry, buf)
	putEntry(entry)
}

func (logger *Logger) writeBuf(entry LogEntry, buf []byte) {
	if WriterService.Running.Load() {
		writeAsync(logPayload{
			entry:  entry,
			buf:    buf,
			writer: logger.writer,
			file:   logger.file,
		})
		return
	}

	if logger.writer != nil {
		logger.writer.Log(entry, buf)
	} else if logger.file != nil {
		fmt.Println(Viewable(string(buf)))
	} else if logger.goLogger == nil {
		fmt.Println(Viewable(string(buf)))
	} else {
		logger.goLogger.Print(string(buf))
	}
}

func (logger *Logger) FillBase(base *BaseLog, logType string) {
	if base == nil {
		return
	}

	base.LogName = logger.config.Name
	if logType != "" {
		base.LogType = logType
	}
	base.LogTime = time.Now().UnixNano()
	base.TraceId = logger.traceId
	if dockerImageTag != "" {
		base.Image = dockerImageName + ":" + dockerImageTag
	} else {
		base.Image = dockerImageName
	}
	if serverIp != "" {
		base.Server = serverName + ":" + serverIp
	} else {
		base.Server = serverName
	}
}

func (logger *Logger) FillDebug(entry *DebugLog, message string) {
	logger.FillBase(&entry.BaseLog, LogTypeDebug)
	entry.Debug = message
}

func (logger *Logger) FillInfo(entry *InfoLog, message string) {
	logger.FillBase(&entry.BaseLog, LogTypeInfo)
	entry.Info = message
}

func (logger *Logger) FillWarning(entry *WarningLog, message string) {
	logger.FillBase(&entry.BaseLog, LogTypeWarning)
	entry.Warning = message
}

func (logger *Logger) FillError(entry *ErrorLog, message string) {
	logger.FillBase(&entry.BaseLog, LogTypeError)
	entry.Error = message
	entry.CallStacks = getCallStacks(logger.truncations)
}

func (logger *Logger) GetCallStacks() []string {
	return getCallStacks(logger.truncations)
}

func (logger *Logger) Debug(message string, extra ...any) {
	if logger.CheckLevel(DEBUG) {
		entry := GetEntry[DebugLog]()
		logger.FillDebug(entry, message)
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Info(message string, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry[InfoLog]()
		logger.FillInfo(entry, message)
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Warning(message string, extra ...any) {
	if logger.CheckLevel(WARNING) {
		entry := GetEntry[WarningLog]()
		logger.FillWarning(entry, message)
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Error(message string, extra ...any) {
	if logger.CheckLevel(ERROR) {
		entry := GetEntry[ErrorLog]()
		logger.FillError(entry, message)
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) SetName(name string) {
	logger.config.Name = name
}

func (logger *Logger) SetLevel(level LevelType) {
	logger.level = level
}

func (logger *Logger) New(traceId string) *Logger {
	newLogger := *logger
	newLogger.traceId = traceId
	return &newLogger
}

func (logger *Logger) GetTraceId() string {
	return logger.traceId
}

// As 仿照 cast.As，忽略错误并返回零值，但会将错误记录到日志中 (消除摩擦)
func (logger *Logger) As(v any, err error) any {
	if err != nil {
		logger.Error(err.Error())
	}
	return v
}

func (logger *Logger) CheckLevel(logLevel LevelType) bool {
	settedLevel := logger.level
	if settedLevel == 0 {
		settedLevel = INFO
	}
	return logLevel >= settedLevel
}
