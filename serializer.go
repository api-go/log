package log

import (
	"bytes"
	"reflect"
	"sort"
	"strconv"

	"apigo.cc/go/cast"
)

type fieldAccessor struct {
	indexPath []int
	name      string
}

var (
	accessorsCache = make(map[string][]fieldAccessor)
)

// getAccessors caches the reflection index paths for the flattened fields.
func getAccessors(logType string, model any) []fieldAccessor {
	metaLock.RLock()
	if acc, ok := accessorsCache[logType]; ok {
		metaLock.RUnlock()
		return acc
	}
	metaLock.RUnlock()

	metaLock.Lock()
	defer metaLock.Unlock()

	// Double check
	if acc, ok := accessorsCache[logType]; ok {
		return acc
	}

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var flatFields []fieldInfo
	flattenStructFields(t, &flatFields, nil)

	// Determine final indices (must match meta.go)
	maxLiteralPos := -1
	var highPosFields []fieldInfo
	for _, f := range flatFields {
		if f.pos < 1000 {
			if f.pos > maxLiteralPos {
				maxLiteralPos = f.pos
			}
		} else {
			highPosFields = append(highPosFields, f)
		}
	}

	// Sort high pos fields by their pos
	sort.Slice(highPosFields, func(i, j int) bool {
		return highPosFields[i].pos < highPosFields[j].pos
	})

	finalPosMap := make(map[string]int)
	for _, f := range flatFields {
		if f.pos < 1000 {
			finalPosMap[f.field.Name] = f.pos
		}
	}
	nextPos := maxLiteralPos + 1
	for _, f := range highPosFields {
		finalPosMap[f.field.Name] = nextPos
		nextPos++
	}

	maxPos := nextPos - 1
	accessors := make([]fieldAccessor, maxPos+1)
	for _, f := range flatFields {
		if f.field.Tag.Get("log") == "-" {
			continue
		}
		realPos := finalPosMap[f.field.Name]
		accessors[realPos] = fieldAccessor{
			indexPath: f.field.Index,
			name:      f.field.Name,
		}
	}

	accessorsCache[logType] = accessors
	return accessors
}

// donot export this function
func Marshal(entry LogEntry, sensitiveKeys []string) []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')

	base := entry.GetBaseLog()
	if base == nil {
		buf.WriteByte(']')
		return buf.Bytes()
	}

	logType := base.LogType
	if logType == "" {
		// Fallback for undefined types
		logType = "undefined"
	}

	accessors := getAccessors(logType, entry)
	v := reflect.ValueOf(entry)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i, acc := range accessors {
		if i > 0 {
			buf.WriteByte(',')
		}

		if acc.indexPath == nil {
			buf.WriteByte('0')
			continue
		}

		fv := v.FieldByIndex(acc.indexPath)
		writeValue(&buf, fv, acc.name, sensitiveKeys)
	}

	buf.WriteByte(']')
	return buf.Bytes()
}

func writeValue(buf *bytes.Buffer, v reflect.Value, fieldName string, sensitiveKeys []string) {
	if !v.IsValid() {
		buf.WriteString("null")
		return
	}

	// Check if this root field should be desensitized
	if len(sensitiveKeys) > 0 {
		fixedName := fixField(fieldName)
		for _, sk := range sensitiveKeys {
			if sk == fixedName {
				buf.WriteString(`"***"`)
				return
			}
		}
	}

	switch v.Kind() {
	case reflect.String:
		writeString(buf, v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf.WriteString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		buf.WriteString(strconv.FormatFloat(v.Float(), 'g', -1, 64))
	case reflect.Bool:
		if v.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	default:
		// Use cast for complex types to ensure deep desensitization
		b, _ := cast.ToJSONDesensitizeBytes(v.Interface(), sensitiveKeys)
		if len(b) == 0 {
			buf.WriteString("null")
		} else {
			buf.Write(b)
		}
	}
}

func writeString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\':
			buf.WriteString(`\\`)
		case '"':
			buf.WriteString(`\"`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			buf.WriteByte(c)
		}
	}
	buf.WriteByte('"')
}
