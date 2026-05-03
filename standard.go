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
	Extra      map[string]interface{}
}

func (b *BaseLog) Reset() {
	b.LogName = ""
	b.LogType = ""
	b.LogTime = ""
	b.TraceId = ""
	if b.Extra == nil {
		b.Extra = make(map[string]interface{}, 8)
	} else {
		clear(b.Extra)
	}
}

func (b *BaseLog) Base() *BaseLog {
	return b
}

type DebugLog struct {
	BaseLog
	Debug string
}

func (d *DebugLog) Reset() {
	d.BaseLog.Reset()
	d.Debug = ""
}

func (d *DebugLog) Base() *BaseLog {
	return &d.BaseLog
}

type InfoLog struct {
	BaseLog
	Info string
}

func (i *InfoLog) Reset() {
	i.BaseLog.Reset()
	i.Info = ""
}

func (i *InfoLog) Base() *BaseLog {
	return &i.BaseLog
}

type WarningLog struct {
	BaseLog
	Warning    string
	CallStacks []string
}

func (w *WarningLog) Reset() {
	w.BaseLog.Reset()
	w.Warning = ""
	w.CallStacks = w.CallStacks[:0]
}

func (w *WarningLog) Base() *BaseLog {
	return &w.BaseLog
}

type ErrorLog struct {
	BaseLog
	Error      string
	CallStacks []string
}

func (e *ErrorLog) Reset() {
	e.BaseLog.Reset()
	e.Error = ""
	e.CallStacks = e.CallStacks[:0]
}

func (e *ErrorLog) Base() *BaseLog {
	return &e.BaseLog
}

type DBLog struct {
	BaseLog
	DbType     string
	Dsn        string
	Query      string
	QueryArgs  string
	UsedTime   float32
	Error      string
	CallStacks []string
}

func (d *DBLog) Reset() {
	d.BaseLog.Reset()
	d.DbType = ""
	d.Dsn = ""
	d.Query = ""
	d.QueryArgs = ""
	d.UsedTime = 0
	d.Error = ""
	d.CallStacks = d.CallStacks[:0]
}

func (d *DBLog) Base() *BaseLog {
	return &d.BaseLog
}

type RequestLog struct {
	BaseLog
	ServerId           string
	App                string
	Node               string
	ClientIp           string
	FromApp            string
	FromNode           string
	UserId             string
	DeviceId           string
	ClientAppName      string
	ClientAppVersion   string
	SessionId          string
	RequestId          string
	Host               string
	Scheme             string
	Proto              string
	AuthLevel          int
	Priority           int
	Method             string
	Path               string
	RequestHeaders     map[string]string
	RequestData        map[string]any
	UsedTime           float32
	ResponseCode       int
	ResponseHeaders    map[string]string
	ResponseDataLength uint
	ResponseData       string
}

func (r *RequestLog) Reset() {
	r.BaseLog.Reset()
	r.ServerId = ""
	r.App = ""
	r.Node = ""
	r.ClientIp = ""
	r.FromApp = ""
	r.FromNode = ""
	r.UserId = ""
	r.DeviceId = ""
	r.ClientAppName = ""
	r.ClientAppVersion = ""
	r.SessionId = ""
	r.RequestId = ""
	r.Host = ""
	r.Scheme = ""
	r.Proto = ""
	r.AuthLevel = 0
	r.Priority = 0
	r.Method = ""
	r.Path = ""

	if r.RequestHeaders == nil {
		r.RequestHeaders = make(map[string]string, 8)
	} else {
		clear(r.RequestHeaders)
	}

	if r.RequestData == nil {
		r.RequestData = make(map[string]any, 8)
	} else {
		clear(r.RequestData)
	}

	r.UsedTime = 0
	r.ResponseCode = 0

	if r.ResponseHeaders == nil {
		r.ResponseHeaders = make(map[string]string, 8)
	} else {
		clear(r.ResponseHeaders)
	}

	r.ResponseDataLength = 0
	r.ResponseData = ""
}

func (r *RequestLog) Base() *BaseLog {
	return &r.BaseLog
}
