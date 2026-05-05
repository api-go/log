package log

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"apigo.cc/go/cast"
)

var (
	dockerImageName string
	dockerImageTag  string
	serverName      string
	serverIp        string
)

func init() {
	dockerImageName = os.Getenv("DOCKER_IMAGE_NAME")
	dockerImageTag = os.Getenv("DOCKER_IMAGE_TAG")
	serverName, _ = os.Hostname()
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, a := range addrs {
			if an, ok := a.(*net.IPNet); ok {
				// 忽略 Docker 私有网段
				if an.IP.IsGlobalUnicast() && !strings.HasPrefix(an.IP.To4().String(), "172.17.") {
					serverIp = an.IP.To4().String()
					break
				}
			}
		}
	}
}

// MakeTime 解析纳秒时间戳或 RFC3339 字符串
func MakeTime(v any) time.Time {
	if ts, ok := cast.ToInt64E(v); ok == nil {
		return time.Unix(0, ts)
	}
	tm, _ := time.Parse(time.RFC3339Nano, cast.String(v))
	return tm
}

// MakeUsedTime 计算消耗时间（毫秒）
func MakeUsedTime(startTime, endTime time.Time) float32 {
	return float32(endTime.UnixNano()-startTime.UnixNano()) / 1e6
}

// ParseBaseLog 解析基础日志行
func ParseBaseLog(line string) *BaseLog {
	pos := strings.IndexByte(line, '{')
	if pos == -1 {
		return ParseBadLog(line)
	}

	l := make(map[string]any)
	err := json.Unmarshal([]byte(line[pos:]), &l)
	if err != nil {
		return ParseBadLog(line)
	}

	baseLog := BaseLog{Extra: make(map[string]any)}
	for k, v := range l {
		lk := strings.ToLower(k)
		switch lk {
		case "logname":
			baseLog.LogName = cast.String(v)
		case "logtype":
			baseLog.LogType = cast.String(v)
		case "logtime":
			baseLog.LogTime = cast.Int64(v)
		case "traceid":
			baseLog.TraceId = cast.String(v)
		case "imagename":
			if baseLog.Image != "" {
				baseLog.Image = cast.String(v) + ":" + baseLog.Image
			} else {
				baseLog.Image = cast.String(v)
			}
		case "imagetag":
			if baseLog.Image != "" {
				baseLog.Image = baseLog.Image + ":" + cast.String(v)
			} else {
				baseLog.Image = cast.String(v)
			}
		case "servername":
			if baseLog.Server != "" {
				baseLog.Server = cast.String(v) + ":" + baseLog.Server
			} else {
				baseLog.Server = cast.String(v)
			}
		case "serverip":
			if baseLog.Server != "" {
				baseLog.Server = baseLog.Server + ":" + cast.String(v)
			} else {
				baseLog.Server = cast.String(v)
			}
		default:
			baseLog.Extra[lk] = v
		}
	}
	return &baseLog
}

// ParseBadLog 解析非 JSON 格式的日志
func ParseBadLog(line string) *BaseLog {
	baseLog := BaseLog{Extra: make(map[string]any)}
	baseLog.LogType = LogTypeUndefined
	if len(line) > 19 && line[19] == ' ' {
		tm, err := time.Parse("2006/01/02 15:04:05", line[0:19])
		if err == nil {
			baseLog.LogTime = tm.UnixNano()
			line = line[20:]
		} else {
			return nil
		}
	} else if len(line) > 26 && line[26] == ' ' {
		tm, err := time.Parse("2006/01/02 15:04:05.000000", line[0:26])
		if err == nil {
			baseLog.LogTime = tm.UnixNano()
			line = line[27:]
		} else {
			return nil
		}
	} else {
		return nil
	}
	baseLog.Extra["info"] = line
	return &baseLog
}

// fixField 格式化字段名（去横线、下划线，小写）
func fixField(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return strings.ToLower(s)
}

// getCallStacks 获取调用栈
func getCallStacks(truncations []string) []string {
	callStacks := make([]string, 0)
	inLogger := true
	for i := 0; i < 50; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if strings.Contains(file, "/go/src/") {
			continue
		}
		// 只有在 logger.go, extra.go 等核心实现文件中的帧才被认为是 "inLogger"
		// 这样可以保留测试文件 (xxx_test.go) 的调用栈
		isLogInternal := (strings.Contains(file, "/log/logger.go") ||
			strings.Contains(file, "/log/utility.go") ||
			strings.Contains(file, "/log/standard.go") ||
			strings.Contains(file, "/log/extra.go"))

		if isLogInternal {
			if inLogger {
				continue
			}
		} else {
			inLogger = false
		}

		if truncations != nil {
			for _, truncation := range truncations {
				if pos := strings.Index(file, truncation); pos != -1 {
					file = file[pos+len(truncation):]
				}
			}
		}
		callStacks = append(callStacks, fmt.Sprintf("%s:%d", file, line))
	}
	return callStacks
}

// GetDefaultName 获取默认应用名称
func GetDefaultName() string {
	name := os.Getenv("DISCOVER_APP")
	if name == "" {
		name = os.Getenv("discover_app")
	}
	if name == "" {
		imageName := os.Getenv("DOCKER_IMAGE_NAME")
		if imageName != "" {
			parts := strings.Split(imageName, "/")
			imageName = parts[len(parts)-1]
			imageName = strings.SplitN(imageName, ":", 2)[0]
			imageName = strings.SplitN(imageName, "#", 2)[0]
			name = imageName
		}
	}
	if name == "" {
		name = path.Base(os.Args[0])
	}
	return name
}
