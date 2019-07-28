package log

import (
	"strings"
)

// 所有支持的日志级别。
const (
	PANIC Level = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG

	printLevel
)

// Level 表示一个已知的日志级别。
type Level int

// String 返回 level 对应的字符串值。
func (l Level) String() string {
	switch l {
	case PANIC:
		return "PANIC"
	case FATAL:
		return "FATAL"
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case printLevel:
		return ""
	default:
		return "<UNKNOWN>"
	}
}

// ParseLevel 将字符串形式的 l 解析成 Level，l 不区分大小写。
// 如果 l 并没有被定义，ok 返回 false。
func ParseLevel(l string) (level Level, ok bool) {
	switch strings.ToUpper(l) {
	case "PANIC":
		return PANIC, true
	case "FATAL":
		return FATAL, true
	case "ERROR":
		return ERROR, true
	case "WARN":
		return WARN, true
	case "INFO":
		return INFO, true
	case "DEBUG", "":
		return DEBUG, true
	}

	return DEBUG, false
}
