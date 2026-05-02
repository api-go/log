package log

import (
	"encoding/json"
	"fmt"
	"math"
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

	var logTime time.Time
	if strings.ContainsRune(b.LogTime, 'T') {
		logTime = MakeTime(b.LogTime)
	} else {
		ft := cast.Float64(b.LogTime)
		ts := int64(math.Floor(ft))
		tns := int64((ft - float64(ts)) * 1e9)
		logTime = time.Unix(ts, tns)
	}

	var outs []string
	t1 := strings.Split(logTime.Format("01-02 15:04:05.000"), " ")
	d := t1[0]
	t := ""
	if len(t1) > 1 {
		t = t1[1]
	}
	t2 := strings.Split(t, ".")
	s := ""
	if len(t2) > 1 {
		s = t2[1]
	}
	t = t2[0]
	outs = append(outs, shell.White(shell.Bold, d+" "+t))
	if s != "" {
		outs = append(outs, shell.White("."+s))
	}
	outs = append(outs, " ", shell.Style(shell.TextWhite, shell.Dim, shell.Underline, b.TraceId))

	level := ""
	levelKey := ""
	for _, k := range []string{"debug", "warning", "error", "info", "Debug", "Warning", "Error", "Info"} {
		if b.Extra[k] != nil {
			level = strings.ToLower(k)
			levelKey = k
			break
		}
	}

	if b.LogType == LogTypeRequest {
		method := cast.String(b.Extra["method"])
		path := cast.String(b.Extra["path"])
		code := cast.Int(b.Extra["responsecode"])
		used := float32(cast.Float64(b.Extra["usedtime"]))

		outs = append(outs, " ", shell.Cyan(shell.Bold, "REQUEST"), " ", shell.Cyan(method), " ", path)
		if code >= 500 {
			outs = append(outs, " ", shell.BRed(cast.String(code)))
		} else if code >= 400 {
			outs = append(outs, " ", shell.BYellow(cast.String(code)))
		} else {
			outs = append(outs, " ", shell.BGreen(cast.String(code)))
		}
		outs = append(outs, " ", shell.Style(shell.Dim, fmt.Sprintf("%.2fms", used)))

		delete(b.Extra, "method")
		delete(b.Extra, "path")
		delete(b.Extra, "responsecode")
		delete(b.Extra, "usedtime")
		delete(b.Extra, "host")
		delete(b.Extra, "scheme")
		delete(b.Extra, "proto")
		delete(b.Extra, "clientip")
		delete(b.Extra, "serverid")
		delete(b.Extra, "app")
		delete(b.Extra, "node")
		delete(b.Extra, "fromapp")
		delete(b.Extra, "fromnode")
		delete(b.Extra, "userid")
		delete(b.Extra, "deviceid")
		delete(b.Extra, "clientappname")
		delete(b.Extra, "clientappversion")
		delete(b.Extra, "sessionid")
		delete(b.Extra, "requestid")
		delete(b.Extra, "authlevel")
		delete(b.Extra, "priority")
		delete(b.Extra, "requestheaders")
		delete(b.Extra, "requestdata")
		delete(b.Extra, "responseheaders")
		delete(b.Extra, "responsedatalength")
		delete(b.Extra, "responsedata")
		delete(b.Extra, "logname")
		delete(b.Extra, "logtype")
		delete(b.Extra, "logtime")
		delete(b.Extra, "traceid")
		delete(b.Extra, "imagename")
		delete(b.Extra, "imagetag")
		delete(b.Extra, "servername")
		delete(b.Extra, "serverip")
	} else if b.LogType == LogTypeStatistic {
		outs = append(outs, " ", shell.Cyan(shell.Bold, "STATISTIC"))
	} else if b.LogType == LogTypeTask {
		outs = append(outs, " ", shell.Cyan(shell.Bold, "TASK"))
	} else {
		if level != "" {
			msg := cast.String(b.Extra[levelKey])
			delete(b.Extra, levelKey)
			switch level {
			case "info":
				outs = append(outs, " ", shell.Cyan(msg))
			case "warning":
				outs = append(outs, " ", shell.Yellow(msg))
			case "error":
				outs = append(outs, " ", shell.Red(msg))
			case "debug":
				outs = append(outs, " ", msg)
			}
		} else if b.LogType == "undefined" {
			outs = append(outs, " ", shell.Style(shell.Dim, "-"))
		} else {
			outs = append(outs, " ", shell.Cyan(shell.Bold, b.LogType))
		}
	}

	callStacks := b.Extra["callStacks"]
	delete(b.Extra, "callStacks")

	if b.Extra != nil {
		for k, v := range b.Extra {
			vStr := cast.String(v)
			if k == "extra" && len(vStr) > 0 && vStr[0] == '{' {
				extra := make(map[string]any)
				_ = json.Unmarshal([]byte(vStr), &extra)
				for k2, v2 := range extra {
					outs = append(outs, "  ", shell.Style(shell.TextWhite, shell.Dim, shell.Italic, k2+":"), cast.String(v2))
				}
			} else {
				outs = append(outs, "  ", shell.Style(shell.TextWhite, shell.Dim, shell.Italic, k+":"), vStr)
			}
		}
	}

	if callStacks != nil {
		var callStacksList []any
		if csStr, ok := callStacks.(string); ok && len(csStr) > 2 && csStr[0] == '[' {
			_ = json.Unmarshal([]byte(csStr), &callStacksList)
		} else if csList, ok := callStacks.([]any); ok {
			callStacksList = csList
		}

		if len(callStacksList) > 0 {
			outs = append(outs, "\n")
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
				outs = append(outs, " ", shell.Style(shell.Dim, v))
				// 简化格式化逻辑
				outs = append(outs, shell.Style(shell.TextWhite, postfix), "\n")
			}
		}
	}
	return strings.Join(outs, "")
}
