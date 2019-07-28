package log

import (
	"fmt"
	"log"
)

const (
	stdLoggerSkipLevel = 2
)

// 从标准库继承而来的标记位。
const (
	StdLoggerLongFile  = log.Llongfile
	StdLoggerShortFile = log.Lshortfile
)

// NewStdLogger 创建一个标准日志。
//
// 这里设置的 skip 是相对标准 Logger 的 Output 方法的级别，skip 为 0 时打印 Output 的调用者位置。
// 由于标准 Logger 并没有将 Output 的 calldepth 传过来，因此 Output 的 calldepth 不再有用。
//
// 这里的 flag 是标准 Logger 支持的标记位。
// 出于公司业务特点，这里仅仅用 flag 确定是否显示文件名，即只有 Llongfile 或 Lshortfile 有用，
// 其他标记位不起作用。
func NewStdLogger(logger Logger, level Level, skip int, flag int) *log.Logger {
	writer := &writerWrapper{
		logger: logger,
		level:  printLevel,
	}
	l := newDefaultLogger(level, writer)
	l.SetSkip(skip + stdLoggerSkipLevel)

	if flag&(StdLoggerLongFile|StdLoggerShortFile) != 0 {
		l.SetShowFileLine(true)
	}

	return log.New(&writerWrapper{
		logger: l,
		level:  level,
	}, "", 0)
}

type writerWrapper struct {
	logger Logger
	level  Level
}

func (w *writerWrapper) Write(p []byte) (n int, err error) {
	s := string(p)

	switch w.level {
	case PANIC:
		w.logger.Panic(s)
	case FATAL:
		w.logger.Fatal(s)
	case ERROR:
		w.logger.Error(s)
	case WARN:
		w.logger.Warn(s)
	case INFO:
		w.logger.Info(s)
	case DEBUG:
		w.logger.Debug(s)
	case printLevel:
		w.logger.Print(s)
	default:
		err = fmt.Errorf("log: invalid log level [level:%v]", w.level)
		return
	}

	n = len(p)
	return
}
