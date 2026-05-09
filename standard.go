package log

const LogTypeDebug = "debug"
const LogTypeInfo = "info"
const LogTypeWarning = "warning"
const LogTypeError = "error"
const LogTypeUndefined = "undefined"
const LogTypeDb = "db"
const LogTypeDbError = "dbError"
const LogTypeServer = "server"
const LogTypeServerError = "serverError"
const LogTypeTask = "task"
const LogTypeMonitor = "monitor"
const LogTypeStatistic = "statistic"
const LogTypeRequest = "request"

const LogDefaultSensitive = "phone,password,secret,token,accessToken"
const LogEnvLevel = "LOG_LEVEL"
const LogEnvFile = "LOG_FILE"
const LogEnvSensitive = "LOG_SENSITIVE"

// LogEntry 是一个标记接口，用于识别是否为对象池管理的日志对象
type LogEntry interface {
	IsLogEntry() bool
	GetBaseLog() *BaseLog
	Reset()
}

type BaseLog struct {
	LogName string         `log:"pos:0,color:cyan,hide:true"`
	LogType string         `log:"pos:1,color:magenta,hide:true"`
	LogTime int64          `log:"pos:2,format:time"`
	TraceId string         `log:"pos:3,color:blue"`
	Image   string         `log:"pos:4,color:darkGray,hide:true"`
	Server  string         `log:"pos:5,color:darkGray,hide:true"`
	Extra   map[string]any `log:"pos:1000"`
}

func (b *BaseLog) IsLogEntry() bool {
	return true
}

func (b *BaseLog) GetBaseLog() *BaseLog {
	return b
}

func (b *BaseLog) Reset() {
	b.LogName = ""
	b.LogType = ""
	b.LogTime = 0
	b.TraceId = ""
	b.Image = ""
	b.Server = ""
	if b.Extra == nil {
		b.Extra = make(map[string]any, 8)
	} else {
		clear(b.Extra)
	}
}

type DebugLog struct {
	BaseLog
	Debug string `log:"pos:6,withoutkey:true"` // white
}

func (l *DebugLog) Reset() {
	l.BaseLog.Reset()
	l.Debug = ""
}

type InfoLog struct {
	BaseLog
	Info string `log:"pos:6,color:cyan,withoutkey:true"`
}

func (l *InfoLog) Reset() {
	l.BaseLog.Reset()
	l.Info = ""
}

type WarningLog struct {
	BaseLog
	Warning    string   `log:"pos:6,color:yellow,withoutkey:true"`
	CallStacks []string `log:"pos:1001"`
}

func (l *WarningLog) Reset() {
	l.BaseLog.Reset()
	l.Warning = ""
	l.CallStacks = l.CallStacks[:0]
}

type ErrorLog struct {
	BaseLog
	Error      string   `log:"pos:6,color:red,withoutkey:true"`
	CallStacks []string `log:"pos:1001"`
}

func (l *ErrorLog) Reset() {
	l.BaseLog.Reset()
	l.Error = ""
	l.CallStacks = l.CallStacks[:0]
}

func init() {
	RegisterType(LogTypeDebug, &DebugLog{})
	RegisterType(LogTypeInfo, &InfoLog{})
	RegisterType(LogTypeWarning, &WarningLog{})
	RegisterType(LogTypeError, &ErrorLog{})
}
