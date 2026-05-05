package log

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"apigo.cc/go/cast"
)

type Logger struct {
	config          Config
	level           LevelType
	goLogger        *log.Logger
	file            *FileWriter
	writer          Writer
	formatter       Formatter
	truncations     []string
	sensitive       map[string]bool
	sensitiveKeys   []string
	regexSensitive  []*regexp.Regexp
	sensitiveRule   []sensitiveRuleInfo
	desensitization func(string) string
	traceId         string
}

type sensitiveRuleInfo struct {
	threshold int
	leftNum   int
	rightNum  int
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
	if conf.SensitiveRule == "" {
		conf.SensitiveRule = "12:4*4, 11:3*4, 7:2*2, 3:1*1, 2:1*0"
	}

	if conf.Name == "" {
		conf.Name = GetDefaultName()
	}

	logger := Logger{
		truncations: cast.Split(conf.Truncations, ","),
		formatter:   conf.Formatter,
	}
	if logger.formatter == nil {
		logger.formatter = &JSONFormatter{}
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

	if len(conf.RegexSensitive) > 0 {
		ss := cast.Split(conf.RegexSensitive, ",")
		for _, v := range ss {
			if r, err := regexp.Compile(v); err == nil {
				logger.regexSensitive = append(logger.regexSensitive, r)
			}
		}
	}

	if len(conf.SensitiveRule) > 0 {
		ss := cast.Split(conf.SensitiveRule, ",")
		for _, v := range ss {
			a1 := strings.SplitN(v, ":", 2)
			if len(a1) == 2 {
				a2 := strings.SplitN(a1[1], "*", 3)
				if len(a2) == 2 {
					threshold := cast.Int(a1[0])
					leftNum := cast.Int(a2[0])
					rightNum := cast.Int(a2[1])
					if threshold >= 0 && threshold <= 100 && leftNum >= 0 && leftNum <= 100 && rightNum >= 0 && rightNum <= 100 {
						logger.sensitiveRule = append(logger.sensitiveRule, sensitiveRuleInfo{
							threshold: threshold,
							leftNum:   leftNum,
							rightNum:  rightNum,
						})
					}
				}
			}
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
	buf, err := logger.formatter.Format(entry, logger.sensitiveKeys)

	if err == nil {
		logger.writeBuf(buf)
	}
	PutEntry(entry)
}

func (logger *Logger) writeBuf(buf []byte) {
	if writerRunning.Load() {
		WriteAsync(logPayload{
			buf:    buf,
			writer: logger.writer,
			file:   logger.file,
		})
		return
	}

	if logger.writer != nil {
		logger.writer.Log(buf)
	} else if logger.file != nil {
		fmt.Println(Viewable(string(buf)))
	} else if logger.goLogger == nil {
		fmt.Println(Viewable(string(buf)))
	} else {
		logger.goLogger.Print(string(buf))
	}
}

func (logger *Logger) fillBase(entry LogEntry, logType string) {
	base := entry.GetBaseLog()
	if base == nil {
		return
	}

	base.LogName = logger.config.Name
	base.LogType = logType
	base.LogTime = time.Now().UnixNano()
	base.TraceId = logger.traceId
	base.ImageName = dockerImageName
	base.ImageTag = dockerImageTag
	base.ServerName = serverName
	base.ServerIp = serverIp
}

func (logger *Logger) Debug(message string, extra ...any) {
	if logger.CheckLevel(DEBUG) {
		entry := GetEntry[DebugLog]()
		logger.fillBase(entry, LogTypeDebug)
		entry.Debug = message
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Info(message string, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry[InfoLog]()
		logger.fillBase(entry, LogTypeInfo)
		entry.Info = message
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Warning(message string, extra ...any) {
	if logger.CheckLevel(WARNING) {
		entry := GetEntry[WarningLog]()
		logger.fillBase(entry, LogTypeWarning)
		entry.Warning = message
		entry.CallStacks = getCallStacks(logger.truncations)
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Error(message string, extra ...any) {
	if logger.CheckLevel(ERROR) {
		entry := GetEntry[ErrorLog]()
		logger.fillBase(entry, LogTypeError)
		entry.Error = message
		entry.CallStacks = getCallStacks(logger.truncations)
		if len(extra) > 0 {
			cast.FillMap(&entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) SetLevel(level LevelType) {
	logger.level = level
}

func (logger *Logger) SetDesensitization(f func(v string) string) {
	logger.desensitization = f
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
