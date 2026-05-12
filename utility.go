package log

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
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

	// 获取真实局域网 IP (UDP 8.8.8.8 伪拨号法)
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		serverIp = localAddr.IP.String()
		_ = conn.Close()
	}

	if serverIp == "" {
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			for _, a := range addrs {
				if an, ok := a.(*net.IPNet); ok {
					if an.IP.IsGlobalUnicast() {
						serverIp = an.IP.To4().String()
						break
					}
				}
			}
		}
	}
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

var globalDefaultName string

// getDefaultName 获取默认应用名称
func getDefaultName() string {
	if globalDefaultName != "" {
		return globalDefaultName
	}
	name := ""
	if info, ok := debug.ReadBuildInfo(); ok && info.Path != "" && info.Path != "command-line-arguments" {
		name = path.Base(info.Path)
	}
	if name == "" {
		name = path.Base(os.Args[0])
	}
	// 处理 Windows 下的 .exe 后缀
	name = strings.TrimSuffix(name, ".exe")
	return name
}
