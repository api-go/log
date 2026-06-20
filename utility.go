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

// isLogPkgFile checks whether a runtime.Caller file path belongs to the log
// package's internal implementation files. It handles both local paths
// (.../log/file.go) and Go module paths with version suffixes
// (.../log@v1.5.6/file.go).
func isLogPkgFile(file, name string) bool {
	return strings.HasSuffix(file, "/log/"+name) ||
		(strings.Contains(file, "/log@") && strings.HasSuffix(file, "/"+name))
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
		// 兼容本地路径 (.../log/file.go) 和 module 路径 (.../log@v1.5.6/file.go)
		isLogInternal := isLogPkgFile(file, "logger.go") ||
			isLogPkgFile(file, "utility.go") ||
			isLogPkgFile(file, "standard.go") ||
			isLogPkgFile(file, "default_logger.go") ||
			isLogPkgFile(file, "extra.go")

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
