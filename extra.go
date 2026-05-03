package log

import (
	"reflect"

	"apigo.cc/go/cast"
)

type TaskLog struct {
	BaseLog
	Task     string
	UsedTime float32
	Success  bool
	Message  string
}

func (t *TaskLog) Reset() {
	t.BaseLog.Reset()
	t.Task = ""
	t.UsedTime = 0
	t.Success = false
	t.Message = ""
}

func (t *TaskLog) Base() *BaseLog {
	return &t.BaseLog
}

type MonitorLog struct {
	BaseLog
	Target  string
	Status  int
	Message string
}

func (m *MonitorLog) Reset() {
	m.BaseLog.Reset()
	m.Target = ""
	m.Status = 0
	m.Message = ""
}

func (m *MonitorLog) Base() *BaseLog {
	return &m.BaseLog
}

type StatisticLog struct {
	BaseLog
	Category string
	Item     string
	Value    float64
}

func (s *StatisticLog) Reset() {
	s.BaseLog.Reset()
	s.Category = ""
	s.Item = ""
	s.Value = 0
}

func (s *StatisticLog) Base() *BaseLog {
	return &s.BaseLog
}

func (logger *Logger) Task(taskName string, usedTime float32, success bool, message string, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry(reflect.TypeOf(&TaskLog{})).(*TaskLog)
		logger.fillBase(entry.Base(), LogTypeTask)
		entry.Task = taskName
		entry.UsedTime = usedTime
		entry.Success = success
		entry.Message = message
		if len(extra) > 0 {
			for i := 0; i < len(extra); i += 2 {
				if i+1 < len(extra) {
					entry.Extra[cast.String(extra[i])] = extra[i+1]
				}
			}
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Monitor(target string, status int, message string, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry(reflect.TypeOf(&MonitorLog{})).(*MonitorLog)
		logger.fillBase(entry.Base(), LogTypeMonitor)
		entry.Target = target
		entry.Status = status
		entry.Message = message
		if len(extra) > 0 {
			for i := 0; i < len(extra); i += 2 {
				if i+1 < len(extra) {
					entry.Extra[cast.String(extra[i])] = extra[i+1]
				}
			}
		}
		logger.Log(entry)
	}
}

func (logger *Logger) Statistic(category, item string, value float64, extra ...any) {
	if logger.CheckLevel(INFO) {
		entry := GetEntry(reflect.TypeOf(&StatisticLog{})).(*StatisticLog)
		logger.fillBase(entry.Base(), LogTypeStatistic)
		entry.Category = category
		entry.Item = item
		entry.Value = value
		if len(extra) > 0 {
			for i := 0; i < len(extra); i += 2 {
				if i+1 < len(extra) {
					entry.Extra[cast.String(extra[i])] = extra[i+1]
				}
			}
		}
		logger.Log(entry)
	}
}
