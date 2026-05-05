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
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Color      string `json:"color,omitempty"`
	Format     string `json:"format,omitempty"`
	WithoutKey bool   `json:"withoutKey,omitempty"`
	Hide       bool   `json:"hide,omitempty"`
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

// fieldInfo is used internally for sorting fields before flattening.
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

	var flatFields []reflect.StructField
	flattenStructFields(t, &flatFields, nil)

	var metaFields []MetaField
	var extraField *reflect.StructField
	var callStacksField *reflect.StructField

	// Process fields, separating Extra and CallStacks
	var regularFields []reflect.StructField
	for _, f := range flatFields {
		if f.Name == "Extra" {
			extraField = &f
			continue
		}
		if f.Name == "CallStacks" {
			callStacksField = &f
			continue
		}
		regularFields = append(regularFields, f)
	}

	// Reassemble: regular fields -> CallStacks -> Extra
	var finalFields []reflect.StructField
	finalFields = append(finalFields, regularFields...)
	if callStacksField != nil {
		finalFields = append(finalFields, *callStacksField)
	}
	if extraField != nil {
		finalFields = append(finalFields, *extraField)
	}

	for i, f := range finalFields {
		tag := f.Tag.Get("log")
		if tag == "-" {
			continue
		}

		meta := MetaField{
			Index: i,
			Name:  f.Name,
		}

		if tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
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
					}
				}
			}
		}
		// Apply some default visual rules if not specified
		// LogType shouldn't show the key in standard console
		if f.Name == "LogType" && meta.Color == "" {
			meta.WithoutKey = true
		}
		
		metaFields = append(metaFields, meta)
	}

	return metaFields
}

func flattenStructFields(t reflect.Type, result *[]reflect.StructField, parentIndex []int) {
	var infos []fieldInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() && !f.Anonymous {
			continue
		}

		isEmbeddedStruct := f.Anonymous && f.Type.Kind() == reflect.Struct
		pos := 1000 + i // default position if not specified
		if isEmbeddedStruct {
			pos = i - 1000 // default to top priority for embedded structs
		}

		tag := f.Tag.Get("log")
		if tag != "" {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				kv := strings.SplitN(part, ":", 2)
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "pos" {
					if p := cast.To[int](strings.TrimSpace(kv[1])); p > 0 {
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

		infos = append(infos, fieldInfo{
			field: f,
			pos:   pos,
		})
	}

	// Sort fields in the current struct level by pos
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].pos < infos[j].pos
	})

	for _, info := range infos {
		if info.field.Anonymous && info.field.Type.Kind() == reflect.Struct {
			// Embedded struct, extract its fields first (parent first)
			flattenStructFields(info.field.Type, result, info.field.Index)
		} else {
			*result = append(*result, info.field)
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
