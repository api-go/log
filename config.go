package log

// Config 日志配置
type Config struct {
	Name           string
	Level          string
	File           string
	Fast           bool
	SplitTag       string
	Truncations    string
	Sensitive      string
	RegexSensitive string
	SensitiveRule  string
	KeepKeyCase    bool // 是否保持Key的首字母大小写？默认一律使用小写
}

type LevelType int

const (
	DEBUG   LevelType = 1
	INFO    LevelType = 2
	WARNING LevelType = 3
	ERROR   LevelType = 4
	CLOSE   LevelType = 5
)
