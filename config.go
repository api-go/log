package log

// Config 日志配置
type Config struct {
	Name        string
	Level       string
	File        string
	SplitTag    string
	Truncations string
	Sensitive   string
}

type LevelType int

const (
	DEBUG   LevelType = 1
	INFO    LevelType = 2
	WARNING LevelType = 3
	ERROR   LevelType = 4
	CLOSE   LevelType = 5
)
