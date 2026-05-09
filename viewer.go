package log

import (
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
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "[") {
		// Fallback highlight for non-array strings
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

	var arr []any
	if err := cast.UnmarshalJSON([]byte(line), &arr); err != nil {
		return line
	}

	if len(arr) < 2 {
		return line
	}

	logType := ""
	if len(arr) > 2 {
		logType = cast.String(arr[2])
	}
	meta := GetMeta(logType)
	if len(meta) == 0 {
		logType = cast.String(arr[1])
		meta = GetMeta(logType)
	}

	if len(meta) == 0 {
		// Fallback rendering
		return fallbackRenderArray(arr)
	}

	var builder strings.Builder

	for i, v := range arr {
		if v == nil || cast.String(v) == "0" { // 0 is gap
			continue
		}
		if i >= len(meta) {
			// Unmapped trailing values, just print them
			builder.WriteString(" ")
			builder.WriteString(shell.Style(shell.Dim, fmt.Sprintf("Index%d:", i)))
			builder.WriteString(cast.String(v))
			continue
		}

		m := meta[i]

		if m.Hide || m.Name == "" {
			continue
		}

		if m.Name == "Extra" {
			extraMap, ok := v.(map[string]any)
			if ok && len(extraMap) > 0 {
				for k, ev := range extraMap {
					builder.WriteString("  ")
					builder.WriteString(shell.Style(shell.TextWhite, shell.Dim, shell.Italic, k+":"))
					vStr := ""
					switch ev.(type) {
					case map[string]any, []any:
						vStr, _ = cast.ToJSON(ev)
					default:
						vStr = cast.String(ev)
					}
					builder.WriteString(vStr)
				}
			}
			continue
		}

		if m.Name == "CallStacks" {
			callStacksList, ok := v.([]any)
			if ok && len(callStacksList) > 0 {
				builder.WriteString("\n")
				for _, vi := range callStacksList {
					vStr := cast.String(vi)
					postfix := ""
					if pos := strings.LastIndexByte(vStr, '/'); pos != -1 {
						postfix = vStr[pos+1:]
						vStr = vStr[:pos+1]
					} else {
						postfix = vStr
						vStr = ""
					}
					builder.WriteString(" ")
					builder.WriteString(shell.Style(shell.Dim, vStr))
					builder.WriteString(shell.Style(shell.TextWhite, postfix))
					builder.WriteString("\n")
				}
			}
			continue
		}

		// Handle normal fields
		vStr := ""
		if m.Format == "time" {
			// Convert int64 ns to time string
			logTime := time.Unix(0, cast.Int64(v))
			vStr = logTime.Format("01-02 15:04:05.000")
			if m.Color == "" {
				builder.WriteString(shell.White(shell.Bold, vStr))
				builder.WriteString(" ")
				continue
			}
		} else {
			vStr = cast.String(v)
			if vStr == "" {
				continue
			}
		}

		if builder.Len() > 0 {
			builder.WriteString(" ")
		}

		if !m.WithoutKey {
			builder.WriteString(shell.Style(shell.TextWhite, shell.Dim, shell.Italic, m.Name+":"))
		}

		builder.WriteString(applyColor(vStr, m.Color))
	}

	return builder.String()
}

// ToJSON converts a JSON array log line to a standard JSON object string based on metadata.
func ToJSON(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "[") {
		return line
	}

	var arr []any
	if err := cast.UnmarshalJSON([]byte(line), &arr); err != nil {
		return line
	}

	if len(arr) < 2 {
		return line
	}

	logType := ""
	if len(arr) > 2 {
		logType = cast.String(arr[2])
	}
	meta := GetMeta(logType)
	if len(meta) == 0 {
		logType = cast.String(arr[1])
		meta = GetMeta(logType)
	}

	if len(meta) == 0 {
		return line
	}

	result := make(map[string]any)
	for i, v := range arr {
		if i < len(meta) {
			m := meta[i]
			if m.Name == "" {
				continue
			}
			if m.Name == "Extra" {
				if extraMap, ok := v.(map[string]any); ok {
					for k, ev := range extraMap {
						result[k] = ev
					}
				}
			} else {
				result[m.Name] = v
			}
		} else if cast.String(v) != "0" {
			result[fmt.Sprintf("Extra%d", i)] = v
		}
	}

	jsonStr, _ := cast.ToJSON(result)
	return jsonStr
}

func applyColor(text string, color string) string {
	switch color {
	case "red":
		return shell.Red(text)
	case "cyan":
		return shell.Cyan(text)
	case "blue":
		return shell.Blue(text)
	case "magenta":
		return shell.Magenta(text)
	case "yellow":
		return shell.Yellow(text)
	case "green":
		return shell.Green(text)
	case "gray", "darkGray":
		return shell.Style(shell.Dim, text)
	default:
		return text
	}
}

func fallbackRenderArray(arr []any) string {
	var builder strings.Builder
	for i, v := range arr {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(cast.String(v))
	}
	return builder.String()
}
