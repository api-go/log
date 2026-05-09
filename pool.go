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
	if le, ok := any(entry).(LogEntry); ok {
		le.Reset()
	}
	return entry
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
