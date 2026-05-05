package log

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"apigo.cc/go/cast"
	"apigo.cc/go/shell"
)

var errorLineMatcher = regexp.MustCompile(`(\w+\.go:\d+)`)
var codeFileMatcher = regexp.MustCompile(`(\w+?\.)(go|js)`)

func Viewable(line string) string {
	b := ParseBaseLog(line)
	if b == nil {
		// 高亮错误代码
		if strings.Contains(line, ".go:") {
			if strings.Contains(line, "/ssgo/") || strings.Contains(line, "/ssdo/") || strings.Contains(line, "/gojs/") {
				line = errorLineMatcher.ReplaceAllString(line, shell.BYellow("$1"))
			} else if !strings.Contains(line, "/apigo.cc/") {
				line = errorLineMatcher.ReplaceAllString(line, shell.BMagenta("$1"))
			} else if !strings.Contains(line, "/go/src/") {
				line = errorLineMatcher.ReplaceAllString(line, shell.BRed("$1"))
			}
		}
		return line
	}

	logTime := time.Unix(0, b.LogTime)

	var builder strings.Builder
	builder.WriteString(shell.White(shell.Bold, logTime.Format("01-02 15:04:05.000")))
	builder.WriteString(" ")
	builder.WriteString(shell.Style(shell.TextWhite, shell.Dim, shell.Underline, b.TraceId))

	level := ""
	for _, k := range []string{"info", "warning", "error", "debug"} {
		if v := b.Extra[k]; v != nil && cast.String(v) != "" {
			level = k
			break
		}
	}

	if b.LogType == LogTypeRequest {
		method := cast.String(b.Extra["method"])
		path := cast.String(b.Extra["path"])
		code := cast.Int(b.Extra["responsecode"])
		used := float32(cast.Float64(b.Extra["usedtime"]))

		builder.WriteString(" ")
		builder.WriteString(shell.Cyan(shell.Bold, "REQUEST"))
		builder.WriteString(" ")
		builder.WriteString(shell.Cyan(method))
		builder.WriteString(" ")
		builder.WriteString(path)
		builder.WriteString(" ")
		if code >= 500 {
			builder.WriteString(shell.BRed(cast.String(code)))
		} else if code >= 400 {
			builder.WriteString(shell.BYellow(cast.String(code)))
		} else {
			builder.WriteString(shell.BGreen(cast.String(code)))
		}
		builder.WriteString(" ")
		builder.WriteString(shell.Style(shell.Dim, fmt.Sprintf("%.2fms", used)))

		for _, k := range []string{"method", "path", "responsecode", "usedtime", "host", "scheme", "proto", "clientip", "serverid", "app", "node", "fromapp", "fromnode", "userid", "deviceid", "clientappname", "clientappversion", "sessionid", "requestid", "authlevel", "priority", "requestheaders", "requestdata", "responseheaders", "responsedatalength", "responsedata", "logname", "logtype", "logtime", "traceid", "imagename", "imagetag", "servername", "serverip"} {
			delete(b.Extra, k)
		}
	} else if b.LogType == LogTypeStatistic {
		builder.WriteString(" ")
		builder.WriteString(shell.Cyan(shell.Bold, "STATISTIC"))
	} else if b.LogType == LogTypeTask {
		builder.WriteString(" ")
		builder.WriteString(shell.Cyan(shell.Bold, "TASK"))
	} else {
		if level != "" {
			msg := cast.String(b.Extra[level])
			delete(b.Extra, level)
			builder.WriteString(" ")
			switch level {
			case "info":
				builder.WriteString(shell.Cyan(msg))
			case "warning":
				builder.WriteString(shell.Yellow(msg))
			case "error":
				builder.WriteString(shell.Red(msg))
			case "debug":
				builder.WriteString(msg)
			}
		} else if b.LogType == "undefined" {
			builder.WriteString(" ")
			builder.WriteString(shell.Style(shell.Dim, "-"))
		} else {
			builder.WriteString(" ")
			builder.WriteString(shell.Cyan(shell.Bold, b.LogType))
		}
	}

	callStacks := b.Extra["callstacks"]
	delete(b.Extra, "callstacks")

	if b.Extra != nil {
		for k, v := range b.Extra {
			vStr := ""
			if v == nil {
				continue
			}
			switch v.(type) {
			case map[string]any, []any:
				vStr, _ = cast.ToJSON(v)
			default:
				vStr = cast.String(v)
			}

			if k == "extra" && len(vStr) > 0 && vStr[0] == '{' {
				extra, err := cast.ToMap[string, any](vStr)
				if err == nil {
					for k2, v2 := range extra {
						builder.WriteString("  ")
						builder.WriteString(shell.Style(shell.TextWhite, shell.Dim, shell.Italic, k2+":"))
						builder.WriteString(cast.String(v2))
					}
				}
			} else {
				builder.WriteString("  ")
				builder.WriteString(shell.Style(shell.TextWhite, shell.Dim, shell.Italic, k+":"))
				builder.WriteString(vStr)
			}
		}
	}

	if callStacks != nil {
		var callStacksList []any
		switch cs := callStacks.(type) {
		case string:
			if len(cs) > 2 && cs[0] == '[' {
				_ = json.Unmarshal([]byte(cs), &callStacksList)
			}
		case []any:
			callStacksList = cs
		}

		if len(callStacksList) > 0 {
			builder.WriteString("\n")
			for _, vi := range callStacksList {
				v := cast.String(vi)
				postfix := ""
				if pos := strings.LastIndexByte(v, '/'); pos != -1 {
					postfix = v[pos+1:]
					v = v[:pos+1]
				} else {
					postfix = v
					v = ""
				}
				builder.WriteString(" ")
				builder.WriteString(shell.Style(shell.Dim, v))
				builder.WriteString(shell.Style(shell.TextWhite, postfix))
				builder.WriteString("\n")
			}
		}
	}
	return builder.String()
}
