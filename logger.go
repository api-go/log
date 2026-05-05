package log

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"apigo.cc/go/cast"
)

type Logger struct {
	config      Config
	level       LevelType
	goLogger    *log.Logger
	file        *FileWriter
	writer      Writer
	truncations []string
	sensitive   map[string]bool
	sensitiveKeys []string
	traceId     string
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
		conf.Name = GetDefaultName()
	}

	logger := Logger{
		truncations: cast.Split(conf.Truncations, ","),
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
					writerLock.Lock()
					cur := writers.Load().([]Writer)
					newW := append(cur, w)
					writers.Store(newW)
					writerLock.Unlock()
					Start()
				}
			}
		} else {
			if conf.SplitTag != "" {
				filesLock.RLock()
				logger.file = files[conf.File+conf.SplitTag]
				filesLock.RUnlock()
				if logger.file == nil {
					logger.file = &FileWriter{
						fileName: conf.File,
						splitTag: conf.SplitTag,
					}
					filesLock.Lock()
					files[conf.File+conf.SplitTag] = logger.file
					filesLock.Unlock()
				}
				Start()
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
	buf := ToArrayBytes(entry, logger.sensitiveKeys)
	logger.writeBuf(entry, buf)
	PutEntry(entry)
}

func (logger *Logger) writeBuf(entry LogEntry, buf []byte) {
	if writerRunning.Load() {
		WriteAsync(logPayload{
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
	entry.CallStacks = getCallStacks(logger.truncations)
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

func (logger *Logger) CheckLevel(logLevel LevelType) bool {
	settedLevel := logger.level
	if settedLevel == 0 {
		settedLevel = INFO
	}
	return logLevel >= settedLevel
}
