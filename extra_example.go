package log

// type RequestLog struct {
// 	BaseLog
// 	ServerId           string            `log:"pos:6"`
// 	App                string            `log:"pos:7"`
// 	Node               string            `log:"pos:8"`
// 	ClientIp           string            `log:"pos:9"`
// 	FromApp            string            `log:"pos:10"`
// 	FromNode           string            `log:"pos:11"`
// 	UserId             string            `log:"pos:12"`
// 	DeviceId           string            `log:"pos:13"`
// 	ClientAppName      string            `log:"pos:14"`
// 	ClientAppVersion   string            `log:"pos:15"`
// 	SessionId          string            `log:"pos:16"`
// 	RequestId          string            `log:"pos:17"`
// 	Host               string            `log:"pos:18"`
// 	Scheme             string            `log:"pos:19"`
// 	Proto              string            `log:"pos:20"`
// 	AuthLevel          int               `log:"pos:21"`
// 	Priority           int               `log:"pos:22"`
// 	Method             string            `log:"pos:23"`
// 	Path               string            `log:"pos:24"`
// 	RequestHeaders     map[string]string `log:"pos:25"`
// 	RequestData        map[string]any    `log:"pos:26"`
// 	UsedTime           float32           `log:"pos:27"`
// 	ResponseCode       int               `log:"pos:28"`
// 	ResponseHeaders    map[string]string `log:"pos:29"`
// 	ResponseDataLength uint              `log:"pos:30"`
// 	ResponseData       string            `log:"pos:31"`
// }

// func (l *RequestLog) Reset() {
// 	l.BaseLog.Reset()
// 	l.ServerId = ""
// 	l.App = ""
// 	l.Node = ""
// 	l.ClientIp = ""
// 	l.FromApp = ""
// 	l.FromNode = ""
// 	l.UserId = ""
// 	l.DeviceId = ""
// 	l.ClientAppName = ""
// 	l.ClientAppVersion = ""
// 	l.SessionId = ""
// 	l.RequestId = ""
// 	l.Host = ""
// 	l.Scheme = ""
// 	l.Proto = ""
// 	l.AuthLevel = 0
// 	l.Priority = 0
// 	l.Method = ""
// 	l.Path = ""
// 	if l.RequestHeaders == nil {
// 		l.RequestHeaders = make(map[string]string, 8)
// 	} else {
// 		clear(l.RequestHeaders)
// 	}
// 	if l.RequestData == nil {
// 		l.RequestData = make(map[string]any, 8)
// 	} else {
// 		clear(l.RequestData)
// 	}
// 	l.UsedTime = 0
// 	l.ResponseCode = 0
// 	if l.ResponseHeaders == nil {
// 		l.ResponseHeaders = make(map[string]string, 8)
// 	} else {
// 		clear(l.ResponseHeaders)
// 	}
// 	l.ResponseDataLength = 0
// 	l.ResponseData = ""
// }

// func (logger *Logger) Request(
// 	method, path, host, scheme, proto string,
// 	clientIp, serverId, app, node string,
// 	fromApp, fromNode string,
// 	userId, deviceId, sessionId, requestId string,
// 	clientAppName, clientAppVersion string,
// 	authLevel, priority int,
// 	reqHeaders map[string]string,
// 	reqData map[string]any,
// 	responseCode int,
// 	usedTime float32,
// 	respHeaders map[string]string,
// 	responseData string,
// 	responseDataLength uint,
// 	extra ...any,
// ) {
// 	if !logger.CheckLevel(INFO) {
// 		return
// 	}

// 	entry := GetEntry[RequestLog]()
// 	logger.FillBase(&entry.BaseLog, LogTypeRequest)

// 	// 暴力平铺赋值，性能极高
// 	entry.Method = method
// 	entry.Path = path
// 	entry.Host = host
// 	entry.Scheme = scheme
// 	entry.Proto = proto
// 	entry.ClientIp = clientIp
// 	entry.ServerId = serverId
// 	entry.App = app
// 	entry.Node = node
// 	entry.FromApp = fromApp
// 	entry.FromNode = fromNode
// 	entry.UserId = userId
// 	entry.DeviceId = deviceId
// 	entry.SessionId = sessionId
// 	entry.RequestId = requestId
// 	entry.ClientAppName = clientAppName
// 	entry.ClientAppVersion = clientAppVersion
// 	entry.AuthLevel = authLevel
// 	entry.Priority = priority
// 	entry.RequestHeaders = reqHeaders
// 	entry.RequestData = reqData
// 	entry.ResponseCode = responseCode
// 	entry.UsedTime = usedTime
// 	entry.ResponseHeaders = respHeaders
// 	entry.ResponseData = responseData
// 	entry.ResponseDataLength = responseDataLength
// 	if len(extra) > 0 {
// 		cast.FillMap(&entry.Extra, extra)
// 	}

// 	logger.Log(entry)
// }

// type TaskLog struct {
// 	BaseLog
// 	Task     string  `log:"pos:6"`
// 	UsedTime float32 `log:"pos:7"`
// 	Success  bool    `log:"pos:8"`
// 	Message  string  `log:"pos:9"`
// }

// func (l *TaskLog) Reset() {
// 	l.BaseLog.Reset()
// 	l.Task = ""
// 	l.UsedTime = 0
// 	l.Success = false
// 	l.Message = ""
// }

// type MonitorLog struct {
// 	BaseLog
// 	Target  string `log:"pos:6"`
// 	Status  int    `log:"pos:7"`
// 	Message string `log:"pos:8"`
// }

// func (l *MonitorLog) Reset() {
// 	l.BaseLog.Reset()
// 	l.Target = ""
// 	l.Status = 0
// 	l.Message = ""
// }

// type StatisticLog struct {
// 	BaseLog
// 	Category string  `log:"pos:6"`
// 	Item     string  `log:"pos:7"`
// 	Value    float64 `log:"pos:8"`
// }

// func (l *StatisticLog) Reset() {
// 	l.BaseLog.Reset()
// 	l.Category = ""
// 	l.Item = ""
// 	l.Value = 0
// }

// func (logger *Logger) Task(taskName string, usedTime float32, success bool, message string, extra ...any) {
// 	if logger.CheckLevel(INFO) {
// 		entry := GetEntry[TaskLog]()
// 		logger.FillBase(&entry.BaseLog, LogTypeTask)
// 		entry.Task = taskName
// 		entry.UsedTime = usedTime
// 		entry.Success = success
// 		entry.Message = message
// 		if len(extra) > 0 {
// 			cast.FillMap(&entry.Extra, extra)
// 		}
// 		logger.Log(entry)
// 	}
// }

// func (logger *Logger) Monitor(target string, status int, message string, extra ...any) {
// 	if logger.CheckLevel(INFO) {
// 		entry := GetEntry[MonitorLog]()
// 		logger.FillBase(&entry.BaseLog, LogTypeMonitor)
// 		entry.Target = target
// 		entry.Status = status
// 		entry.Message = message
// 		if len(extra) > 0 {
// 			cast.FillMap(&entry.Extra, extra)
// 		}
// 		logger.Log(entry)
// 	}
// }

// func (logger *Logger) Statistic(category, item string, value float64, extra ...any) {
// 	if logger.CheckLevel(INFO) {
// 		entry := GetEntry[StatisticLog]()
// 		logger.FillBase(&entry.BaseLog, LogTypeStatistic)
// 		entry.Category = category
// 		entry.Item = item
// 		entry.Value = value
// 		if len(extra) > 0 {
// 			cast.FillMap(&entry.Extra, extra)
// 		}
// 		logger.Log(entry)
// 	}
// }

// func init() {
// 	RegisterType(LogTypeRequest, &RequestLog{})
// 	RegisterType(LogTypeTask, &TaskLog{})
// 	RegisterType(LogTypeMonitor, &MonitorLog{})
// 	RegisterType(LogTypeStatistic, &StatisticLog{})
// }
