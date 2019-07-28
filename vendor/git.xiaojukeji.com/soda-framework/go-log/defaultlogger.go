package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"git.xiaojukeji.com/soda-framework/go-log/internal/lumberjack"
)

const (
	defaultLoggerTimeFormat        = "2006-01-02T15:04:05.000-0700"
	defaultLoggerQueueSize         = 1 << 15
	defaultLoggerFormatArgsLogSize = 1 << 10
	defaultLoggerFullLogSize       = 1 << 11
	defaultLoggerSkipLevel         = 3
	maxLoggerEntriesInOneWrite     = 32              // 用来尽量减少文件写入次数，这是打包写入日志文件的数量。
	defaultLoggerMinRotateInterval = 2 * time.Second // 用来定期检查日志文件是否还存在，不存在就 rotate，精度不应该太高，否则会影响服务器性能。

	deviceNull   = os.DevNull
	deviceStdout = "/dev/stdout"
	deviceStderr = "/dev/stderr"

	vendorPath = "/vendor/"
)

var (
	newLine = []byte("\n")

	// 用来测试的时间戳，如果设置了就会使用这个值而不是 entry.logTime。
	defaultLoggerTestTime *time.Time
)

func init() {
	lumberjack.BackupTimeFormat = "20060102150405"
}

type levelWriter struct {
	maxLevel Level
	writer   io.WriteCloser
}

type defaultLogger struct {
	addWriter chan levelWriter
	queue     chan *entry
	close     chan struct{}
	exit      chan struct{}

	maxLevel     Level
	showFileLine bool
	callerSkip   int
	formatter    string

	closed int32
}

var (
	_ Logger         = new(defaultLogger)
	_ Skipper        = new(defaultLogger)
	_ io.WriteCloser = new(defaultLogger)
)

func newDefaultLogger(maxLevel Level, writer io.Writer) *defaultLogger {
	l := &defaultLogger{
		addWriter: make(chan levelWriter),
		queue:     make(chan *entry, defaultLoggerQueueSize),
		close:     make(chan struct{}),
		exit:      make(chan struct{}),

		maxLevel: maxLevel,
	}
	go l.loop()
	l.AddWriter(printLevel, writer)
	return l
}

func (l *defaultLogger) AddWriter(maxLevel Level, writer io.Writer) {
	wc, ok := writer.(io.WriteCloser)

	if !ok {
		wc = nopWriteCloser{writer}
	}

	l.addWriter <- levelWriter{
		maxLevel: maxLevel,
		writer:   wc,
	}
}

func (l *defaultLogger) SetShowFileLine(enabled bool) {
	l.showFileLine = enabled
}

func (l *defaultLogger) SetFormatter(formatter string) {
	l.formatter = strings.ToLower(formatter)
}

func (l *defaultLogger) SetSkip(skip int) {
	l.callerSkip = skip
}

func (l *defaultLogger) SetMaxLevel(maxLevel Level) {
	l.maxLevel = maxLevel
}

func (l *defaultLogger) Close() error {
	// 避免重复关闭。
	if !atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		return nil
	}

	close(l.close)
	select {
	case <-l.exit:
	}

	return nil
}

func (l *defaultLogger) Printf(format string, args ...interface{}) {
	l.logf(printLevel, format, args...)
}

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	l.logf(DEBUG, format, args...)
}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.logf(INFO, format, args...)
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	l.logf(WARN, format, args...)
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
}

func (l *defaultLogger) Fatalf(format string, args ...interface{}) {
	l.logf(FATAL, format, args...)
}

func (l *defaultLogger) Panicf(format string, args ...interface{}) {
	l.logf(PANIC, format, args...)
}

func (l *defaultLogger) Print(args ...interface{}) {
	l.log(printLevel, args...)
}

func (l *defaultLogger) Debug(args ...interface{}) {
	l.log(DEBUG, args...)
}

func (l *defaultLogger) Info(args ...interface{}) {
	l.log(INFO, args...)
}

func (l *defaultLogger) Warn(args ...interface{}) {
	l.log(WARN, args...)
}

func (l *defaultLogger) Error(args ...interface{}) {
	l.log(ERROR, args...)
}

func (l *defaultLogger) Fatal(args ...interface{}) {
	l.log(FATAL, args...)
}

func (l *defaultLogger) Panic(args ...interface{}) {
	l.log(PANIC, args...)
}

func (l *defaultLogger) Write(p []byte) (n int, err error) {
	n = len(p)
	formatArgs := getFormatArgs()
	formatArgs.buf = bytes.NewBuffer(p)
	formatArgs.unmanaged = true
	l.enqueue(printLevel, formatArgs)
	return
}

func (l *defaultLogger) logf(level Level, format string, args ...interface{}) {
	if l.maxLevel < level && level != printLevel {
		return
	}

	formatArgs := getFormatArgs()
	formatArgs.format = format
	formatArgs.args = args
	formatArgs.useFormat = true
	l.enqueue(level, formatArgs)
}

func (l *defaultLogger) log(level Level, args ...interface{}) {
	if l.maxLevel < level && level != printLevel {
		return
	}

	formatArgs := getFormatArgs()
	formatArgs.args = args
	l.enqueue(level, formatArgs)
}

func (l *defaultLogger) enqueue(level Level, formatArgs *logFormatArgs) {
	if atomic.LoadInt32(&l.closed) > 0 {
		return
	}

	var pc uintptr

	if level != printLevel {
		if l.showFileLine {
			if p, _, _, ok := runtime.Caller(defaultLoggerSkipLevel + l.callerSkip); ok {
				pc = p
			}
		}
	}

	// formatArgs 里面的 args 可能包含 map 或指针，
	// 如果延时打印就可能造成打出的数据过旧或者发生并发冲突，
	// 所以必须在这里就做完格式化的操作。
	msg := formatArgs.Format()

	// 入队列。
	now := time.Now()
	entry := getEntry()
	entry.level = level
	entry.logTime = now
	entry.formatArgs = formatArgs
	entry.pc = pc
	entry.useJSON = l.formatter == "json"
	l.queue <- entry

	if level == FATAL {
		Flush()
		os.Exit(1)
		return
	}

	if level == PANIC {
		Flush()
		panic(string(msg))
	}
}

func (l *defaultLogger) loop() {
	var writers []levelWriter
	bufs := make([]*bytes.Buffer, maxLoggerEntriesInOneWrite)
	allEntries := make([]*entry, maxLoggerEntriesInOneWrite)

	for i := 0; i < maxLoggerEntriesInOneWrite; i++ {
		buf := &bytes.Buffer{}
		buf.Grow(defaultLoggerFullLogSize)
		bufs[i] = buf
	}

	for {
		select {
		case entry := <-l.queue:
			entries := allEntries[:0:maxLoggerEntriesInOneWrite]
			entry.buf = bufs[0]
			entry.buf.Reset()
			entries = append(entries, entry)

		DrainRemaining:
			for i := 1; i < maxLoggerEntriesInOneWrite; i++ {
				select {
				case entry := <-l.queue:
					entry.buf = bufs[i]
					entry.buf.Reset()
					entries = append(entries, entry)
				default:
					break DrainRemaining
				}
			}

			write(writers, entries)
		case lw := <-l.addWriter:
			writers = append(writers, lw)
		case <-l.close:
			entries := allEntries[:0:maxLoggerEntriesInOneWrite]

		DumpRemaining:
			for {
				select {
				case entry := <-l.queue:
					i := len(entries)

					if i < maxLoggerEntriesInOneWrite {
						entry.buf = bufs[i]
						entry.buf.Reset()
					} else {
						entry.buf = &bytes.Buffer{}
						entry.buf.Grow(defaultLoggerFullLogSize)
					}

					entries = append(entries, entry)
				default:
					break DumpRemaining
				}
			}

			write(writers, entries)

			for _, wc := range writers {
				wc.writer.Close()
			}

			close(l.exit)
			return
		}
	}
}

func write(writers []levelWriter, entries []*entry) {
	for _, entry := range entries {
		buf := entry.buf

		var file, name string
		var line int

		if entry.pc > 0 {
			f := runtime.FuncForPC(entry.pc)
			file, line = f.FileLine(entry.pc)
			name = f.Name()

			if idx := strings.LastIndex(name, vendorPath); idx >= 0 {
				name = name[idx+len(vendorPath):]
			}
		}

		if entry.useJSON {
			msg := entry.formatArgs.Format()
			encoder := json.NewEncoder(buf)
			logTime := entry.logTime

			if defaultLoggerTestTime != nil {
				logTime = *defaultLoggerTestTime
			}

			jl := jsonLog{
				Level: entry.level.String(),
				Time:  logTime.Format(defaultLoggerTimeFormat),
				Msg:   *(*string)(unsafe.Pointer(&msg)),
			}

			if entry.pc > 0 {
				jl.File = fmt.Sprintf("%v:%v@%v", path.Base(file), line, name)
			}

			encoder.Encode(jl)
		} else {
			if entry.level != printLevel {
				// [level][time]
				buf.WriteRune('[')
				buf.WriteString(entry.level.String())
				buf.WriteRune(']')
				buf.WriteRune('[')
				buf.WriteString(entry.logTime.Format(defaultLoggerTimeFormat))
				buf.WriteRune(']')

				// [file:line@funcName]
				if entry.pc > 0 {
					buf.WriteRune('[')
					buf.WriteString(path.Base(file))
					buf.WriteRune(':')
					buf.WriteString(strconv.Itoa(line))
					buf.WriteRune('@')
					buf.WriteString(name)
					buf.WriteRune(']')
				}

				buf.WriteRune(' ')
			}

			buf.Write(entry.formatArgs.Format())
			buf.Write(newLine)
		}
	}

	for _, w := range writers {
		for _, entry := range entries {
			if entry.level <= w.maxLevel {
				w.writer.Write(entry.buf.Bytes())
			}
		}
	}

	for i, entry := range entries {
		entries[i] = nil
		putEntry(entry)
	}
}

type defaultLoggerConfigurer struct {
	config       *Config
	globalLogger *defaultLogger
	fileOutput   io.WriteCloser
	errorOutput  io.WriteCloser

	loggersMu      sync.Mutex
	exit           chan struct{}
	closed         chan struct{}
	createdLoggers map[string]*lumberjack.Logger
}

func newDefaultLoggerConfigurer(config *Config) *defaultLoggerConfigurer {
	dc := &defaultLoggerConfigurer{
		config: config,

		exit:           make(chan struct{}, 1),
		closed:         make(chan struct{}),
		createdLoggers: map[string]*lumberjack.Logger{},
	}

	dc.init()
	go dc.autoRotateLogs()
	return dc
}

func (dc *defaultLoggerConfigurer) init() {
	config := dc.config

	if config.FilePath != "" {
		dc.fileOutput = dc.newLumberjack(config.FilePath, config.MaxSizeMB, config.MaxBackups)
	}

	if config.ErrorFilePath != "" {
		if config.ErrorFilePath == config.FilePath {
			dc.errorOutput = nopWriteCloser{dc.fileOutput}
		} else {
			dc.errorOutput = dc.newLumberjack(config.ErrorFilePath, config.MaxSizeMB, config.MaxBackups)
		}
	}
}

func (dc *defaultLoggerConfigurer) Configure(filePath string) Logger {
	output := dc.fileOutput

	// 对于跟默认日志同名的日志，需要留意不要重复关闭输出。
	if filePath == dc.config.FilePath {
		output = nopWriteCloser{output}
	} else if filePath != "" {
		// 对于一般的文件日志，需要创建新的文件。
		output = dc.newLumberjack(filePath, dc.config.MaxSizeMB, dc.config.MaxBackups)
	}

	level, ok := ParseLevel(dc.config.Level)

	if !ok {
		level = DEBUG
	}

	logger := newDefaultLogger(level, output)
	logger.SetShowFileLine(dc.config.ShowFileLine)

	errorOutput := dc.errorOutput

	// 只有默认日志需要关闭错误日志文件，其他的日志仅需要使用。
	if filePath != "" {
		errorOutput = nopWriteCloser{errorOutput}
	}

	logger.AddWriter(ERROR, errorOutput)

	// 在调试模式下，所有日志同时输出到 stderr。
	if dc.config.Debug {
		logger.AddWriter(level, nopWriteCloser{os.Stderr})
	}

	return logger
}

func (dc *defaultLoggerConfigurer) newLumberjack(filename string, maxSizeMB int, maxBackups int) io.WriteCloser {
	if filename == deviceNull {
		return nopWriteCloser{ioutil.Discard}
	} else if filename == deviceStdout {
		return nopWriteCloser{os.Stdout}
	} else if filename == deviceStderr {
		return nopWriteCloser{os.Stderr}
	}

	logger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		LocalTime:  true,
	}

	if absPath, err := filepath.Abs(filename); err == nil {
		dc.loggersMu.Lock()
		defer dc.loggersMu.Unlock()

		if l, ok := dc.createdLoggers[absPath]; ok {
			return l
		}

		dc.createdLoggers[absPath] = logger
	}

	return logger
}

func (dc *defaultLoggerConfigurer) Close() error {
	dc.loggersMu.Lock()
	dc.createdLoggers = map[string]*lumberjack.Logger{}
	dc.loggersMu.Unlock()

	dc.exit <- struct{}{}
	select {
	case <-dc.closed:
	}
	return nil
}

// autoRotateLogs 可以以 defaultLoggerMinRotateInterval 的频率自动检查日志是否需要 rotate。
func (dc *defaultLoggerConfigurer) autoRotateLogs() {
	ticker := time.NewTicker(defaultLoggerMinRotateInterval)

	for {
		select {
		case <-dc.exit:
			ticker.Stop()
			close(dc.closed)
			return
		case <-ticker.C:
			dc.tryRotateLogs()
		}
	}
}

func (dc *defaultLoggerConfigurer) tryRotateLogs() {
	dc.loggersMu.Lock()
	defer dc.loggersMu.Unlock()

	for absPath, logger := range dc.createdLoggers {
		info, err := os.Stat(absPath)

		if os.IsNotExist(err) {
			logger.Close()
			continue
		}

		if f := logger.File(); f != nil {
			fileInfo, err := f.Stat()

			// 为什么会无法 stat 一个打开了的文件？
			// 一般来说只可能是这个文件已经不存在了，所以这里直接做了 rotate。
			if err != nil {
				logger.Close()
				continue
			}

			// 打开的文件已经不是配置中的文件，重新 rotate。
			if !os.SameFile(info, fileInfo) {
				logger.Close()
				continue
			}
		}
	}
}

func (dc *defaultLoggerConfigurer) sync() {
	dc.loggersMu.Lock()
	defer dc.loggersMu.Unlock()

	for _, logger := range dc.createdLoggers {
		logger.Close()
	}
}

type nopWriteCloser struct {
	Writer io.Writer
}

func (wc nopWriteCloser) Write(data []byte) (int, error) {
	return wc.Writer.Write(data)
}

func (wc nopWriteCloser) Close() error {
	return nil
}

type entry struct {
	buf        *bytes.Buffer
	level      Level
	logTime    time.Time
	formatArgs *logFormatArgs
	pc         uintptr
	useJSON    bool
}

var (
	entryPool = &sync.Pool{
		New: func() interface{} {
			return &entry{}
		},
	}
	zeroEntry entry
)

func getEntry() *entry {
	return entryPool.Get().(*entry)
}

func putEntry(entry *entry) {
	putFormatArgs(entry.formatArgs)
	*entry = zeroEntry
	entryPool.Put(entry)
}

type jsonLog struct {
	Level string `json:"level,omitempty"`
	Time  string `json:"time"`
	Msg   string `json:"msg"`
	File  string `json:"file,omitempty"`
}

type logFormatArgs struct {
	buf *bytes.Buffer

	format    string
	args      []interface{}
	useFormat bool
	unmanaged bool
}

var (
	logFormatArgsPool = &sync.Pool{
		New: func() interface{} {
			return &logFormatArgs{}
		},
	}
	zeroFormatArgs logFormatArgs

	bufPool = &sync.Pool{
		New: func() interface{} {
			buf := &bytes.Buffer{}
			buf.Grow(defaultLoggerFormatArgsLogSize)
			return buf
		},
	}
)

func getFormatArgsBuf() *bytes.Buffer {
	return bufPool.Get().(*bytes.Buffer)
}

func putFormatArgsBuf(buf *bytes.Buffer) {
	buf.Reset()
	bufPool.Put(buf)
}

func getFormatArgs() *logFormatArgs {
	return logFormatArgsPool.Get().(*logFormatArgs)
}

func putFormatArgs(formatArgs *logFormatArgs) {
	if !formatArgs.unmanaged && formatArgs.buf != nil {
		putFormatArgsBuf(formatArgs.buf)
	}

	*formatArgs = zeroFormatArgs
	logFormatArgsPool.Put(formatArgs)
}

func (formatArgs *logFormatArgs) Format() []byte {
	if formatArgs.buf != nil {
		return formatArgs.buf.Bytes()
	}

	buf := getFormatArgsBuf()

	if formatArgs.useFormat {
		fmt.Fprintf(buf, formatArgs.format, formatArgs.args...)
	} else {
		args := formatArgs.args

		if len(args) > 0 {
			fmt.Fprint(buf, args[0])
			args = args[1:]
		}

		for _, arg := range args {
			buf.WriteString("||")
			fmt.Fprint(buf, arg)
		}
	}

	formatArgs.buf = buf
	return formatArgs.buf.Bytes()
}
