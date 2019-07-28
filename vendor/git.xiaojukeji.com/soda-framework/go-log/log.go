// Package log 定义了一个通用的日志组件，其内部的日志实现可以被替换成任意的实现。
package log

import (
	"io"
	"os"
	"sync"
)

// Logger 是一个通用的日志输出接口，
type Logger interface {
	// 需要有关闭日志的能力。
	io.Closer

	// 一系列带格式的日志函数。
	Printf(format string, args ...interface{}) // 无视日志级别，始终都能输出日志，内容里不包含任何额外信息。
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{}) // 应该调用 os.Exit 终止程序。
	Panicf(format string, args ...interface{}) // 应该调用 panic 终止程序。

	// 一系列不带格式的日志函数。
	Print(args ...interface{}) // 无视日志级别，始终都能输出日志，内容里不包含任何额外信息。
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{}) // 应该调用 os.Exit 终止程序。
	Panic(args ...interface{}) // 应该调用 panic 终止程序。
}

// Configurer 在 New 的时候用来初始化 Logger 细节。
// 由于所有创建的 Logger 实例最终都会调用 Close() 关闭，
// Configer 不需要手动关闭打开的 Logger，只需要确保释放自己管理的资源，
// 比如额外的一些监控用 goroutine 等。
type Configurer interface {
	io.Closer
	Configure(filePath string) Logger
}

// Skipper 用于设置 call stack 的 skip level。
// 当 skip 为 0 时 Logger 会显示调用者的 call stack；
// 当 skip > 0 时 Logger 会显示调用者再往前回溯 skip 层的 call stack。
type Skipper interface {
	SetSkip(skip int) // 设置新的 skip level。
}

var (
	globalLogger                 = New("").(*delegateLogger)
	globalConfigurer             = defaultConfigurer
	defaultConfigurer Configurer = &globalConfigure{
		creator: func(string) io.WriteCloser { return nopWriteCloser{os.Stderr} },
	}

	createdLoggersMu sync.Mutex
	createdLoggers   = map[string]*delegateLogger{}
)

// New 创建一个新的 Logger。
func New(filePath string) Logger {
	createdLoggersMu.Lock()
	defer createdLoggersMu.Unlock()

	if l, ok := createdLoggers[filePath]; ok {
		return l
	}

	l := &delegateLogger{}
	l.replace(globalConfigurer.Configure(filePath))
	createdLoggers[filePath] = l
	return l
}

// Init 初始化默认 Logger 的配置。如果使用第三方日志库，可以不调用这个函数。
func Init(config *Config) {
	if config.FilePath == "" {
		config.FilePath = DefaultFilePath
	}

	if config.ErrorFilePath == "" {
		config.ErrorFilePath = DefaultErrorFilePath
	}

	if config.Level == "" {
		config.Level = DefaultLevel
	}

	if config.MaxSizeMB <= 0 {
		config.MaxSizeMB = defaultMaxSizeMB
	}

	if config.MaxBackups <= 0 {
		config.MaxBackups = 0
	}

	if config.Formatter == "" {
		config.Formatter = DefaultFormatter
	}

	Use(newDefaultLoggerConfigurer(config))
}

// DefaultLogger 返回默认 Logger。
func DefaultLogger() Logger {
	return globalLogger
}

// Use 设置 Logger 的生成器。
// 一旦 Configer 发生变化，所有已经 New 的日志实例都会重新初始化。
func Use(configurer Configurer) {
	if globalConfigurer == configurer {
		return
	}

	<-use(configurer, false)
}

func use(configurer Configurer, reconfigure bool) <-chan struct{} {
	createdLoggersMu.Lock()
	defer createdLoggersMu.Unlock()

	done := make(chan struct{}, 1)

	if !reconfigure && globalConfigurer == configurer {
		done <- struct{}{}
		return done
	}

	oldConfigurer := globalConfigurer
	globalConfigurer = configurer
	oldLogs := make([]Logger, 0, len(createdLoggers))

	for filePath, delegate := range createdLoggers {
		oldLogs = append(oldLogs, delegate.replace(globalConfigurer.Configure(filePath)))
	}

	go func() {
		oldConfigurer.Close()

		for _, l := range oldLogs {
			l.Close()
		}

		done <- struct{}{}
	}()

	return done
}

// Printf 输出日志，无视当前设置的日志级别。
func Printf(format string, args ...interface{}) {
	globalLogger.logger.Printf(format, args...)
}

// Debugf 输出 DEBUG 级别的日志信息。
func Debugf(format string, args ...interface{}) {
	globalLogger.logger.Debugf(format, args...)
}

// Infof 输出 INFO 级别的日志信息。
func Infof(format string, args ...interface{}) {
	globalLogger.logger.Infof(format, args...)
}

// Warnf 输出 WARN 级别的日志信息。
func Warnf(format string, args ...interface{}) {
	globalLogger.logger.Warnf(format, args...)
}

// Errorf 输出 ERROR 级别的日志信息。
func Errorf(format string, args ...interface{}) {
	globalLogger.logger.Errorf(format, args...)
}

// Fatalf 输出 FATAL 级别的日志信息并采用 os.Exit 退出程序。
func Fatalf(format string, args ...interface{}) {
	globalLogger.logger.Fatalf(format, args...)
}

// Panicf 输出 PANIC 级别的日志信息并采用 panic 抛出异常。
func Panicf(format string, args ...interface{}) {
	globalLogger.logger.Panicf(format, args...)
}

// Print 输出日志，无视当前设置的日志级别。
func Print(args ...interface{}) {
	globalLogger.logger.Print(args...)
}

// Debug 输出 DEBUG 级别的日志信息。
func Debug(args ...interface{}) {
	globalLogger.logger.Debug(args...)
}

// Info 输出 INFO 级别的日志信息。
func Info(args ...interface{}) {
	globalLogger.logger.Info(args...)
}

// Warn 输出 WARN 级别的日志信息。
func Warn(args ...interface{}) {
	globalLogger.logger.Warn(args...)
}

// Error 输出 ERROR 级别的日志信息。
func Error(args ...interface{}) {
	globalLogger.logger.Error(args...)
}

// Fatal 输出 FATAL 级别的日志信息并采用 os.Exit 退出程序。
func Fatal(args ...interface{}) {
	globalLogger.logger.Fatal(args...)
}

// Panic 输出 PANIC 级别的日志信息并采用 panic 抛出异常。
func Panic(args ...interface{}) {
	globalLogger.logger.Panic(args...)
}

// Close 关闭当前日志，如果希望重新使用各种 log 接口，
// 必须使用 Use 设置新的打开的日志组件。
func Close() error {
	<-use(defaultConfigurer, true)
	return nil
}

// Flush 调用后将当前所有未输出的日志输出。
func Flush() {
	<-use(globalConfigurer, true)
}

type globalConfigure struct {
	creator func(filePath string) io.WriteCloser
}

func (gc *globalConfigure) Configure(filePath string) Logger {
	l := newDefaultLogger(DEBUG, gc.creator(filePath))
	l.SetShowFileLine(true)
	return l
}

func (gc *globalConfigure) Close() error {
	return nil
}
