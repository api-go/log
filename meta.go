package log

import (
	"reflect"
	"sort"
	"strings"
	"sync"

	"apigo.cc/go/cast"
	"apigo.cc/go/file"
)

// MetaField describes the serialization and visualization metadata for a single log field.
type MetaField struct {
	Index        int
	Name         string
	KeyName      string
	AttachBefore bool
	Color        string
	Format       string
	Precision    int
	WithoutKey   bool
	Hide         bool
}

var (
	metaRegistry = make(map[string][]MetaField)
	metaLock     sync.RWMutex
	metaFilePath = ".log.meta.json"
)

// LoadMeta loads metadata from the specified file into the global registry.
func LoadMeta(path string) error {
	metaLock.Lock()
	defer metaLock.Unlock()
	return file.UnmarshalFile(path, &metaRegistry)
}

// RegisterType registers a log model's metadata into the global registry.
// logType is the string identifier (e.g. "info", "error").
func RegisterType(logType string, model any) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 强制检查 Reset 方法是否被显式实现（防止继承 BaseLog 后忘记重置业务字段）
	if t.Kind() == reflect.Struct {
		ptrType := reflect.PointerTo(t)
		method, ok := ptrType.MethodByName("Reset")
		if !ok {
			panic("log model must implement Reset() method: " + t.Name())
		}

		// 检查该方法是否属于当前类型（而不是继承自 BaseLog 且没有被重写）
		baseResetMethod, _ := reflect.PointerTo(reflect.TypeOf(BaseLog{})).MethodByName("Reset")
		if method.Func.Pointer() == baseResetMethod.Func.Pointer() {
			panic("log model must override Reset() method to clear its own fields: " + t.Name())
		}
	}

	fields := extractMetaFields(model)

	metaLock.Lock()
	metaRegistry[logType] = fields
	metaLock.Unlock()

	syncMetaFile()
}

// GetMeta returns the metadata fields for a given logType.
func GetMeta(logType string) []MetaField {
	metaLock.RLock()
	defer metaLock.RUnlock()
	return metaRegistry[logType]
}

// fieldInfo is used internally for storing fields with their absolute position.
type fieldInfo struct {
	field reflect.StructField
	pos   int
}

func extractMetaFields(model any) []MetaField {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	var flatFields []fieldInfo
	flattenStructFields(t, &flatFields, nil)

	// Determine final indices
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

	// Assign real indices to high pos fields
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
	metaFields := make([]MetaField, maxPos+1)
	// Initialize with empty MetaFields having Index set
	for i := range metaFields {
		metaFields[i] = MetaField{Index: i}
	}

	for _, f := range flatFields {
		tag := f.field.Tag.Get("log")
		if tag == "-" {
			continue
		}

		realPos := finalPosMap[f.field.Name]
		meta := MetaField{
			Index: realPos,
			Name:  f.field.Name,
		}

		if tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "attachBefore" {
					meta.AttachBefore = true
					continue
				}
				kv := strings.SplitN(part, ":", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					val := strings.TrimSpace(kv[1])
					switch key {
					case "color":
						meta.Color = val
					case "format":
						meta.Format = val
					case "withoutkey":
						meta.WithoutKey = (val == "true")
					case "hide":
						meta.Hide = (val == "true")
					case "keyname":
						meta.KeyName = val
					case "attachBefore":
						meta.AttachBefore = (val == "true")
					case "precision":
						meta.Precision = cast.To[int](val)
					}
				}
			}
		}
		// Apply some default visual rules if not specified
		// LogType shouldn't show the key in standard console
		if f.field.Name == "LogType" && meta.Color == "" {
			meta.WithoutKey = true
		}

		metaFields[realPos] = meta
	}

	return metaFields
}

func flattenStructFields(t reflect.Type, result *[]fieldInfo, parentIndex []int) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() && !f.Anonymous {
			continue
		}

		pos := 10 + i // default position if not specified
		tag := f.Tag.Get("log")
		if tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				kv := strings.SplitN(part, ":", 2)
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "pos" {
					if p := cast.To[int](strings.TrimSpace(kv[1])); p >= 0 {
						pos = p
					}
				}
			}
		}

		// Compute the full index path from the root
		fullIndex := make([]int, len(parentIndex), len(parentIndex)+1)
		copy(fullIndex, parentIndex)
		fullIndex = append(fullIndex, i)
		f.Index = fullIndex

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			flattenStructFields(f.Type, result, f.Index)
		} else {
			*result = append(*result, fieldInfo{
				field: f,
				pos:   pos,
			})
		}
	}
}

func syncMetaFile() {
	metaLock.RLock()
	defer metaLock.RUnlock()
	_ = file.MarshalFilePretty(metaFilePath, metaRegistry)
}

// SetMetaFilePath allows changing the path for testing or configuration purposes
func SetMetaFilePath(path string) {
	metaFilePath = path
}
