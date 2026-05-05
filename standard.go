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
const LogEnvRegexSensitive = "LOG_REGEXSENSITIVE"

// LogEntry 是一个标记接口，用于识别是否为对象池管理的日志对象
type LogEntry interface {
	IsLogEntry() bool
	GetBaseLog() *BaseLog
}

type BaseLog struct {
	LogName string         `log:"pos:1,color:cyan,hide:true"`
	LogType string         `log:"pos:2,color:magenta,hide:true"`
	LogTime int64          `log:"pos:3,format:time"`
	TraceId string         `log:"pos:4,color:blue"`
	Image   string         `log:"pos:5,color:darkGray,hide:true"`
	Server  string         `log:"pos:6,color:darkGray,hide:true"`
	Extra   map[string]any `log:"pos:99"`
}

func (b *BaseLog) IsLogEntry() bool {
	return true
}

func (b *BaseLog) GetBaseLog() *BaseLog {
	return b
}

type DebugLog struct {
	BaseLog
	Debug string `log:"pos:9,withoutkey:true"` // white
}

type InfoLog struct {
	BaseLog
	Info string `log:"pos:9,color:cyan,withoutkey:true"`
}

type WarningLog struct {
	BaseLog
	Warning    string   `log:"pos:9,color:yellow,withoutkey:true"`
	CallStacks []string `log:"pos:98"`
}

type ErrorLog struct {
	BaseLog
	Error      string   `log:"pos:9,color:red,withoutkey:true"`
	CallStacks []string `log:"pos:98"`
}

func init() {
	RegisterType(LogTypeDebug, DebugLog{})
	RegisterType(LogTypeInfo, InfoLog{})
	RegisterType(LogTypeWarning, WarningLog{})
	RegisterType(LogTypeError, ErrorLog{})
}
