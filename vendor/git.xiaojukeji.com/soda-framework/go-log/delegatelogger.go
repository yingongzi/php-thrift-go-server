package log

const (
	delegateLoggerSkipLevel = 1
)

type delegateLogger struct {
	skip   int
	logger Logger
}

func (logger *delegateLogger) replace(l Logger) Logger {
	old := logger.logger
	logger.logger = l
	logger.SetSkip(logger.skip)
	return old
}

// 需要有关闭日志的能力。
func (logger *delegateLogger) Close() error {
	return logger.logger.Close()
}

func (logger *delegateLogger) Printf(fmt string, args ...interface{}) {
	logger.logger.Printf(fmt, args...)
}

func (logger *delegateLogger) Debugf(fmt string, args ...interface{}) {
	logger.logger.Debugf(fmt, args...)
}

func (logger *delegateLogger) Infof(fmt string, args ...interface{}) {
	logger.logger.Infof(fmt, args...)
}

func (logger *delegateLogger) Warnf(fmt string, args ...interface{}) {
	logger.logger.Warnf(fmt, args...)
}

func (logger *delegateLogger) Errorf(fmt string, args ...interface{}) {
	logger.logger.Errorf(fmt, args...)
}

func (logger *delegateLogger) Fatalf(fmt string, args ...interface{}) {
	logger.logger.Fatalf(fmt, args...)
}

func (logger *delegateLogger) Panicf(fmt string, args ...interface{}) {
	logger.logger.Panicf(fmt, args...)
}

func (logger *delegateLogger) Print(args ...interface{}) {
	logger.logger.Print(args...)
}

func (logger *delegateLogger) Debug(args ...interface{}) {
	logger.logger.Debug(args...)
}

func (logger *delegateLogger) Info(args ...interface{}) {
	logger.logger.Info(args...)
}

func (logger *delegateLogger) Warn(args ...interface{}) {
	logger.logger.Warn(args...)
}

func (logger *delegateLogger) Error(args ...interface{}) {
	logger.logger.Error(args...)
}

func (logger *delegateLogger) Fatal(args ...interface{}) {
	logger.logger.Fatal(args...)
}

func (logger *delegateLogger) Panic(args ...interface{}) {
	logger.logger.Panic(args...)
}

func (logger *delegateLogger) SetSkip(skip int) {
	logger.skip = skip

	if skipper, ok := logger.logger.(Skipper); ok {
		skipper.SetSkip(skip + delegateLoggerSkipLevel)
	}
}
