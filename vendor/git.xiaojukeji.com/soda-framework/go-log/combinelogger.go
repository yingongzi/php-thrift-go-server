package log

type combineLogger struct {
	loggers []*delegateLogger
}

// Combine 将多个 logger 合并成一个，内容会复写到多个文件。
// 这个函数只接受由 New 创建出来的 Logger，或者由 Combine 生成的 Logger，
// 对于其他的 Logger 会直接忽略。
func Combine(logger ...Logger) Logger {
	loggers := make([]*delegateLogger, 0, len(logger))

	for _, l := range logger {
		if cl, ok := l.(*combineLogger); ok {
			loggers = append(loggers, cl.loggers...)
		} else if dl, ok := l.(*delegateLogger); ok {
			loggers = append(loggers, dl)
		}
	}

	return &combineLogger{
		loggers: loggers,
	}
}

// Close 不做任何事情，考虑到所有 Logger 肯定都已经被托管，最终会被释放掉，
// 这里不考虑释放的问题。
func (logger *combineLogger) Close() error {
	return nil
}

func (logger *combineLogger) Printf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Printf(format, args...)
	}
}

func (logger *combineLogger) Debugf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Debugf(format, args...)
	}
}

func (logger *combineLogger) Infof(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Infof(format, args...)
	}
}

func (logger *combineLogger) Warnf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Warnf(format, args...)
	}
}

func (logger *combineLogger) Errorf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Errorf(format, args...)
	}
}

func (logger *combineLogger) Fatalf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Fatalf(format, args...)
	}
}

func (logger *combineLogger) Panicf(format string, args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Panicf(format, args...)
	}
}

func (logger *combineLogger) Print(args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Print(args...)
	}
}

func (logger *combineLogger) Debug(args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Debug(args...)
	}
}

func (logger *combineLogger) Info(args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Info(args...)
	}
}

func (logger *combineLogger) Warn(args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Warn(args...)
	}
}

func (logger *combineLogger) Error(args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Error(args...)
	}
}

func (logger *combineLogger) Fatal(args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Fatal(args...)
	}
}

func (logger *combineLogger) Panic(args ...interface{}) {
	for _, l := range logger.loggers {
		l.logger.Panic(args...)
	}
}
