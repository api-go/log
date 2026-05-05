package log

import (
	"reflect"
	"sync"
)

// PoolManager 管理不同日志类型的对象池
type PoolManager struct {
	pools sync.Map // map[reflect.Type]*sync.Pool
}

var (
	globalPools = &PoolManager{}
	resetCache  sync.Map // map[reflect.Type]func(reflect.Value)
)

// GetEntry 从池中获取一个指定类型的日志对象，并确保其处于 Reset 后的干净状态
func GetEntry[T any]() *T {
	t := reflect.TypeFor[*T]()
	p, ok := globalPools.pools.Load(t)
	if !ok {
		p, _ = globalPools.pools.LoadOrStore(t, &sync.Pool{
			New: func() any {
				return new(T)
			},
		})
	}
	entry := p.(*sync.Pool).Get().(*T)
	ResetLogEntry(entry) // 自动重置所有字段，无需子类实现 Reset
	return entry
}

// ResetLogEntry 使用反射自动化重置日志对象的所有字段
// 特别是对 Map 和 Slice 进行初始化（长度0，容量8）
func ResetLogEntry(v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	
	t := rv.Type()
	resetFunc, ok := resetCache.Load(t)
	if !ok {
		resetFunc = buildResetFunc(t.Elem())
		resetCache.Store(t, resetFunc)
	}
	resetFunc.(func(reflect.Value))(rv.Elem())
}

func buildResetFunc(t reflect.Type) func(reflect.Value) {
	var funcs []func(reflect.Value)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldIdx := i
		
		switch field.Type.Kind() {
		case reflect.String:
			funcs = append(funcs, func(rv reflect.Value) { rv.Field(fieldIdx).SetString("") })
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			funcs = append(funcs, func(rv reflect.Value) { rv.Field(fieldIdx).SetInt(0) })
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			funcs = append(funcs, func(rv reflect.Value) { rv.Field(fieldIdx).SetUint(0) })
		case reflect.Float32, reflect.Float64:
			funcs = append(funcs, func(rv reflect.Value) { rv.Field(fieldIdx).SetFloat(0) })
		case reflect.Bool:
			funcs = append(funcs, func(rv reflect.Value) { rv.Field(fieldIdx).SetBool(false) })
		case reflect.Map:
			funcs = append(funcs, func(rv reflect.Value) {
				f := rv.Field(fieldIdx)
				if f.IsNil() {
					f.Set(reflect.MakeMapWithSize(f.Type(), 8))
				} else {
					f.Clear()
				}
			})
		case reflect.Slice:
			funcs = append(funcs, func(rv reflect.Value) {
				f := rv.Field(fieldIdx)
				if f.Cap() < 8 {
					f.Set(reflect.MakeSlice(f.Type(), 0, 8))
				} else {
					f.SetLen(0)
				}
			})
		case reflect.Struct:
			subReset := buildResetFunc(field.Type)
			funcs = append(funcs, func(rv reflect.Value) {
				subReset(rv.Field(fieldIdx))
			})
		case reflect.Ptr, reflect.Interface:
			zero := reflect.Zero(field.Type)
			funcs = append(funcs, func(rv reflect.Value) {
				rv.Field(fieldIdx).Set(zero)
			})
		}
	}
	
	return func(rv reflect.Value) {
		for _, f := range funcs {
			f(rv)
		}
	}
}

func resetStruct(rv reflect.Value) {
	// 已经不再直接调用，保留 buildResetFunc 逻辑即可
}

// PutEntry 将日志对象归还到池中
func PutEntry(entry any) {
	t := reflect.TypeOf(entry)
	if pool, ok := globalPools.pools.Load(t); ok {
		pool.(*sync.Pool).Put(entry)
	}
}

// WithEntry 执行闭包并在结束后自动回收对象
func WithEntry[T any](fn func(*T)) {
	entry := GetEntry[T]()
	defer PutEntry(entry)
	fn(entry)
}

// LogEntry 是一个标记接口，用于识别是否为对象池管理的日志对象
// 已移至 standard.go
