package log

// import (
// 	"apigo.cc/go/cast"
// )

// type RequestLog struct {
// 	BaseLog
// 	ServerId           string
// 	App                string
// 	Node               string
// 	ClientIp           string
// 	FromApp            string
// 	FromNode           string
// 	UserId             string
// 	DeviceId           string
// 	ClientAppName      string
// 	ClientAppVersion   string
// 	SessionId          string
// 	RequestId          string
// 	Host               string
// 	Scheme             string
// 	Proto              string
// 	AuthLevel          int
// 	Priority           int
// 	Method             string
// 	Path               string
// 	RequestHeaders     map[string]string
// 	RequestData        map[string]any
// 	UsedTime           float32
// 	ResponseCode       int
// 	ResponseHeaders    map[string]string
// 	ResponseDataLength uint
// 	ResponseData       string
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
// 	logger.fillBase(entry, LogTypeRequest)

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
// 	Task     string
// 	UsedTime float32
// 	Success  bool
// 	Message  string
// }

// type MonitorLog struct {
// 	BaseLog
// 	Target  string
// 	Status  int
// 	Message string
// }

// type StatisticLog struct {
// 	BaseLog
// 	Category string
// 	Item     string
// 	Value    float64
// }

// func (logger *Logger) Task(taskName string, usedTime float32, success bool, message string, extra ...any) {
// 	if logger.CheckLevel(INFO) {
// 		entry := GetEntry[TaskLog]()
// 		logger.fillBase(entry, LogTypeTask)
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
// 		logger.fillBase(entry, LogTypeMonitor)
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
// 		logger.fillBase(entry, LogTypeStatistic)
// 		entry.Category = category
// 		entry.Item = item
// 		entry.Value = value
// 		if len(extra) > 0 {
// 			cast.FillMap(&entry.Extra, extra)
// 		}
// 		logger.Log(entry)
// 	}
// }
