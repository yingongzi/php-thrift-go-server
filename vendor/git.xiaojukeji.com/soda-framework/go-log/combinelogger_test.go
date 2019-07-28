package log

import (
	"io"
	"strings"
	"testing"
)

func TestCombineLogger(t *testing.T) {
	for i := 0; i < 1000; i++ {
		testCombineLogger(t, i)
	}
}

func testCombineLogger(t *testing.T, round int) {
	var combined Logger
	var bufs []*closableBuffer
	var loggers []*defaultLogger

	swapCreatedLoggers(func() {
		Use(&globalConfigure{
			creator: func(string) io.WriteCloser {
				buf := &closableBuffer{}
				bufs = append(bufs, buf)
				return buf
			},
		})
		l1 := New("")
		l2 := New("foo")
		l3 := New("bar")
		combined = Combine(l1, l2)
		combined = Combine(combined, l3)

		// 并不会真正关闭日志。
		combined.Close()

		combined.Print("print", "a")
		combined.Debug("debug", "b")
		combined.Info("info", "c")
		combined.Warn("warn", "d")
		combined.Error("error", "e")
		mustPanic(t, func() { combined.Panic("panic", "f") })
		combined.Printf("%v, %v", "printf", "a")
		combined.Debugf("%v, %v", "debugf", "b")
		combined.Infof("%v, %v", "infof", "c")
		combined.Warnf("%v, %v", "warnf", "d")
		combined.Errorf("%v, %v", "errorf", "e")
		mustPanic(t, func() { combined.Panicf("%v, %v", "panicf", "f") })

		for _, logger := range combined.(*combineLogger).loggers {
			if dl, ok := logger.logger.(*defaultLogger); ok {
				loggers = append(loggers, dl)
			} else {
				t.Fatalf("logger must be a defaultLogger. [logger:%T]", logger.logger)
			}
		}
	})

	if expected := 3; len(combined.(*combineLogger).loggers) != expected {
		t.Fatalf("invalid logger count. [expected:%v] [actual:%v]", expected, combined.(*combineLogger).loggers)
	}

	lines := make([]string, 0, len(bufs))

	for i, buf := range bufs {
		lines = append(lines, buf.String())

		if !buf.closed {
			t.Fatalf("buf is not closed. [idx:%v]", i)
		}
	}

	if expected, actual := `[DEBUG][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] debugf, b
[DEBUG][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] debugf, b
[DEBUG][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] debugf, b
[DEBUG][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] debug||b
[DEBUG][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] debug||b
[DEBUG][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] debug||b
[ERROR][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] errorf, e
[ERROR][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] errorf, e
[ERROR][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] errorf, e
[ERROR][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] error||e
[ERROR][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] error||e
[ERROR][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] error||e
[INFO][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] infof, c
[INFO][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] infof, c
[INFO][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] infof, c
[INFO][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] info||c
[INFO][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] info||c
[INFO][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] info||c
[PANIC][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1.2] panic||f
[PANIC][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1.3] panicf, f
[WARN][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] warnf, d
[WARN][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] warnf, d
[WARN][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] warnf, d
[WARN][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] warn||d
[WARN][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] warn||d
[WARN][][combinelogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.testCombineLogger.func1] warn||d
printf, a
printf, a
printf, a
print||a
print||a
print||a`, sortLines(maskDateAndFileLine(strings.Join(lines, "\n"))); expected != actual {
		t.Fatalf("invalid log output at round %v.\nexpected:\n%v\nactual:\n%v", round, expected, actual)
	}

	for _, logger := range loggers {
		select {
		case <-logger.close:
		default:
			t.Fatalf("logger must be closed. [logger:%v]", logger)
		}

		if logger.closed != 1 {
			t.Fatalf("logger must be closed. [logger:%v] [closed:%v]", logger, logger.closed)
		}
	}
}

func mustPanic(t *testing.T, fn func()) {
	defer func() {
		e := recover()

		if e == nil {
			t.Fatalf("logger must panic.")
		}
	}()

	fn()
}
