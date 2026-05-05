package log

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
