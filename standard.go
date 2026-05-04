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

type BaseLog struct {
	LogName    string
	LogType    string
	LogTime    string
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

type ErrorLog struct {
	BaseLog
	Error      string
	CallStacks []string
}
