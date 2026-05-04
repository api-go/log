package log

import (
	"reflect"

	"apigo.cc/go/cast"
)

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

func (logger *Logger) Request(entry *RequestLog) {
	logger.fillBase(entry, LogTypeRequest)
	logger.Log(entry)
}

type TaskLog struct {
	BaseLog
	Task     string
	UsedTime float32
	Success  bool
	Message  string
}

type MonitorLog struct {
	BaseLog
	Target  string
	Status  int
	Message string
}

type StatisticLog struct {
	BaseLog
	Category string
	Item     string
	Value    float64
}

func (logger *Logger) Task(taskName string, usedTime float32, success bool, message string, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry(reflect.TypeOf(&TaskLog{})).(*TaskLog)
		logger.fillBase(entry, LogTypeTask)
		entry.Task = taskName
		entry.UsedTime = usedTime
		entry.Success = success
		entry.Message = message
		if len(extra) > 0 {
			cast.ToMap(entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Monitor(target string, status int, message string, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry(reflect.TypeOf(&MonitorLog{})).(*MonitorLog)
		logger.fillBase(entry, LogTypeMonitor)
		entry.Target = target
		entry.Status = status
		entry.Message = message
		if len(extra) > 0 {
			cast.ToMap(entry.Extra, extra)
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Statistic(category, item string, value float64, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry(reflect.TypeOf(&StatisticLog{})).(*StatisticLog)
		logger.fillBase(entry, LogTypeStatistic)
		entry.Category = category
		entry.Item = item
		entry.Value = value
		if len(extra) > 0 {
			cast.ToMap(entry.Extra, extra)
		}
		logger.Log(entry)
	}
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

func (logger *Logger) DB(dbType, dsn, query string, args []any, usedTime float32, err error, extra ...any) {
	logType := LogTypeDb
	level := INFO
	var e string
	if err != nil {
		logType = LogTypeDbError
		level = ERROR
		e = err.Error()
	}

	if logger.CheckLevel(level) {
		entry := GetEntry(reflect.TypeOf(&DBLog{})).(*DBLog)
		logger.fillBase(entry, logType)
		entry.DbType = dbType
		entry.Dsn = dsn
		entry.Query = query
		entry.QueryArgs = cast.MustToJSON(args)
		entry.UsedTime = usedTime
		if e != "" {
			entry.Error = e
			entry.CallStacks = getCallStacks(logger.truncations)
		}
		if len(extra) > 0 {
			cast.ToMap(entry.Extra, extra)
		}
		logger.Log(entry)
	}
}
