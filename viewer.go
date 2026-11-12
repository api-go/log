package log

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/api-go/cast"
	"github.com/api-go/shell"
)

var errorLineMatcher = regexp.MustCompile(`(\w+\.go:\d+)`)
var codeFileMatcher = regexp.MustCompile(`(\w+?\.)(go|js)`)
var workspaceRoot, _ = os.Getwd()

func Viewable(line string) string { return viewable(line, GetMeta) }

func viewable(line string, getMeta func(string) []MetaField) string {
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
	meta := getMeta(logType)
	if len(meta) == 0 {
		logType = cast.String(arr[1])
		meta = getMeta(logType)
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
				keys := make([]string, 0, len(extraMap))
				for k := range extraMap {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					ev := extraMap[k]
					builder.WriteString("  ")
					builder.WriteString(shell.Style(shell.TextWhite, shell.Dim, shell.Italic, k+":"))
					builder.WriteString(renderValue(ev, 0, ""))
				}
			}
			continue
		}

		if m.Name == "CallStacks" {
			callStacksList, ok := v.([]any)
			if ok && len(callStacksList) > 0 {
				stackColor := shell.TextRed
				if strings.Contains(strings.ToLower(logType), "warn") {
					stackColor = shell.TextYellow
				}

				builder.WriteString("\n")
				for _, vi := range callStacksList {
					vStr := cast.String(vi)
					if workspaceRoot != "" {
						vStr = strings.TrimPrefix(vStr, workspaceRoot)
						vStr = strings.TrimPrefix(vStr, "/")
					}
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
					builder.WriteString(shell.Style(stackColor, postfix))
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
			dateStr := logTime.Format("01-02")
			timeStr := logTime.Format("15:04:05")
			milliStr := logTime.Format(".000")

			if builder.Len() > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(shell.Style(shell.Dim, dateStr))
			builder.WriteString(" ")
			builder.WriteString(shell.White(shell.Bold, timeStr))
			builder.WriteString(shell.Style(shell.Dim, milliStr))
			continue
		} else {
			vStr = renderValue(v, m.Precision, m.Color)
			if vStr == "" {
				continue
			}
		}

		if builder.Len() > 0 {
			if m.AttachBefore {
				builder.WriteString(":")
			} else {
				builder.WriteString(" ")
			}
		}

		if !m.WithoutKey && !m.AttachBefore {
			name := m.KeyName
			if name == "" {
				name = m.Name
			}
			builder.WriteString(shell.Style(shell.TextWhite, shell.Dim, shell.Italic, name+":"))
		}

		builder.WriteString(vStr)
	}

	return builder.String()
}

// ToJSON converts a JSON array log line to a standard JSON object string based on metadata.
func ToJSON(line string) string { return toJSON(line, GetMeta) }

func toJSON(line string, getMeta func(string) []MetaField) string {
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
	meta := getMeta(logType)
	if len(meta) == 0 {
		logType = cast.String(arr[1])
		meta = getMeta(logType)
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

			name := m.KeyName
			if name == "" {
				name = m.Name
			}

			if m.Name == "Extra" {
				if extraMap, ok := v.(map[string]any); ok {
					for k, ev := range extraMap {
						result[k] = ev
					}
				}
			} else {
				result[name] = v
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
	case "gray":
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

func renderValue(v any, precision int, color string) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case float32, float64:
		vStr := ""
		if precision > 0 {
			vStr = fmt.Sprintf("%.*f", precision, cast.To[float64](v))
		} else {
			vStr = cast.String(v)
		}
		return applyColor(vStr, color)
	case map[string]any:
		if len(val) == 0 {
			return ""
		}
		var parts []string
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			// Key is always dim, value is colored
			vStr := renderValue(val[k], precision, color)
			if vStr != "" {
				parts = append(parts, fmt.Sprintf("%s:%s", shell.Style(shell.Dim, k), vStr))
			}
		}
		if len(parts) == 0 {
			return ""
		}
		return "[ " + strings.Join(parts, " ") + " ]"
	case []any:
		if len(val) == 0 {
			return ""
		}
		var parts []string
		for _, iv := range val {
			vStr := renderValue(iv, precision, color)
			if vStr != "" {
				parts = append(parts, vStr)
			}
		}
		if len(parts) == 0 {
			return ""
		}
		return "[ " + strings.Join(parts, " ") + " ]"
	default:
		s := cast.String(v)
		if s == "" {
			return ""
		}
		return applyColor(s, color)
	}
}
