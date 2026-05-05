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
	LogName    string
	LogType    string
	LogTime    int64
	TraceId    string
	ImageName  string
	ImageTag   string
	ServerName string
	ServerIp   string
	Extra      map[string]any
}

func (b *BaseLog) IsLogEntry() bool {
	return true
}

func (b *BaseLog) GetBaseLog() *BaseLog {
	return b
}

type StackTraceable interface {
	GetCallStacks() []string
	SetCallStacks([]string)
}

type DebugLog struct {
	BaseLog
	Debug string
}

type InfoLog struct {
	BaseLog
	Info string
}

type WarningLog struct {
	BaseLog
	Warning    string
	CallStacks []string
}

func (l *WarningLog) GetCallStacks() []string { return l.CallStacks }
func (l *WarningLog) SetCallStacks(s []string) { l.CallStacks = s }

type ErrorLog struct {
	BaseLog
	Error      string
	CallStacks []string
}

func (l *ErrorLog) GetCallStacks() []string { return l.CallStacks }
func (l *ErrorLog) SetCallStacks(s []string) { l.CallStacks = s }
