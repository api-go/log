package log

import (
	"reflect"
	"sync"
)

// LogEntry 定义了高性能日志必须实现的接口
type LogEntry interface {
	Reset()
	Base() *BaseLog
}

// PoolManager 管理不同日志类型的对象池
type PoolManager struct {
	pools sync.Map // map[reflect.Type]*sync.Pool
}

var globalPools = &PoolManager{}

// GetEntry 从池中获取一个指定类型的日志对象，并确保其处于 Reset 后的干净状态
func GetEntry(t reflect.Type) LogEntry {
	pool, _ := globalPools.pools.LoadOrStore(t, &sync.Pool{
		New: func() any {
			return reflect.New(t.Elem()).Interface()
		},
	})
	entry := pool.(*sync.Pool).Get().(LogEntry)
	entry.Reset() // 确保获取到的对象永远是干净且预分配好的
	return entry
}

// PutEntry 将日志对象归还到池中，不再进行 Reset
func PutEntry(entry LogEntry) {
	t := reflect.TypeOf(entry)
	if pool, ok := globalPools.pools.Load(t); ok {
		pool.(*sync.Pool).Put(entry)
	}
}

// WithEntry 执行闭包并在结束后自动回收对象
func WithEntry(t reflect.Type, fn func(LogEntry)) {
	entry := GetEntry(t)
	defer PutEntry(entry)
	fn(entry)
}
