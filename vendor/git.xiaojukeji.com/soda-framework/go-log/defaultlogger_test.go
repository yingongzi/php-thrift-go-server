package log

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestDefaultLogger(t *testing.T) {
	c := make(chan bool)
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	l := newDefaultLogger(INFO, buf1)
	l.AddWriter(ERROR, buf2)
	l.AddWriter(DEBUG, nopWriteCloser{os.Stderr})
	l.SetShowFileLine(true)

	l.Printf("printf %v", 1)
	l.Debugf("debugf %v", 2)
	l.Infof("infof %v", 3)
	l.Warnf("warnf %v", 4)
	l.Errorf("errorf %v", 5)

	go call(c, func() {
		callPanicf(l)
	})

	if !<-c {
		t.Fatalf("fail to call Panicf.")
	}

	l.Print("print", 1)
	l.Debug("debug", 2)
	l.Info("info", 3)
	l.Warn("warn", 4)
	l.Error("error", 5)

	go call(c, func() {
		callPanic(l)
	})

	if !<-c {
		t.Fatalf("fail to call Panicf.")
	}

	l.Close()

	actual1 := maskDateAndFileLine(buf1.String())
	actual2 := maskDateAndFileLine(buf2.String())

	if expected := `printf 1
[INFO][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] infof 3
[WARN][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] warnf 4
[ERROR][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] errorf 5
[PANIC][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanicf] panicf 6
print||1
[INFO][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] info||3
[WARN][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] warn||4
[ERROR][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] error||5
[PANIC][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanic] panic||6
`; expected != actual1 {
		t.Fatalf("invalid buf1.\nexpected:\n%v\nactual:\n%v", expected, actual1)
	}

	if expected := `[ERROR][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] errorf 5
[PANIC][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanicf] panicf 6
[ERROR][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLogger] error||5
[PANIC][][defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanic] panic||6
`; expected != actual2 {
		t.Fatalf("invalid buf2.\nexpected:\n%v\nactual:\n%v", expected, actual2)
	}
}

func TestDefaultLoggerWithJSON(t *testing.T) {
	timeString := "2018-01-02T12:34:56.789+0800"
	fakeTime, _ := time.Parse(defaultLoggerTimeFormat, timeString)
	defaultLoggerTestTime = &fakeTime
	defer func() {
		defaultLoggerTestTime = nil
	}()

	c := make(chan bool)
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	l := newDefaultLogger(INFO, buf1)
	l.AddWriter(ERROR, buf2)
	l.AddWriter(DEBUG, nopWriteCloser{os.Stderr})
	l.SetShowFileLine(true)
	l.SetFormatter("json")

	l.Printf("printf %v", 1)
	l.Debugf("debugf %v", 2)
	l.Infof("infof %v", 3)
	l.Warnf("warnf %v", 4)
	l.Errorf("errorf %v", 5)

	go call(c, func() {
		callPanicf(l)
	})

	if !<-c {
		t.Fatalf("fail to call Panicf.")
	}

	l.Print("print", 1)
	l.Debug("debug", 2)
	l.Info("info", 3)
	l.Warn("warn", 4)
	l.Error("error", 5)

	go call(c, func() {
		callPanic(l)
	})

	if !<-c {
		t.Fatalf("fail to call Panicf.")
	}

	l.Close()

	actual1 := maskDateAndFileLine(buf1.String())
	actual2 := maskDateAndFileLine(buf2.String())

	if expected := `{"time":"2018-01-02T12:34:56.789+0800","msg":"printf 1"}
{"level":"INFO","time":"2018-01-02T12:34:56.789+0800","msg":"infof 3","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"WARN","time":"2018-01-02T12:34:56.789+0800","msg":"warnf 4","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"ERROR","time":"2018-01-02T12:34:56.789+0800","msg":"errorf 5","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"PANIC","time":"2018-01-02T12:34:56.789+0800","msg":"panicf 6","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanicf"}
{"time":"2018-01-02T12:34:56.789+0800","msg":"print||1"}
{"level":"INFO","time":"2018-01-02T12:34:56.789+0800","msg":"info||3","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"WARN","time":"2018-01-02T12:34:56.789+0800","msg":"warn||4","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"ERROR","time":"2018-01-02T12:34:56.789+0800","msg":"error||5","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"PANIC","time":"2018-01-02T12:34:56.789+0800","msg":"panic||6","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanic"}
`; expected != actual1 {
		t.Fatalf("invalid buf1.\nexpected:\n%v\nactual:\n%v", expected, actual1)
	}

	if expected := `{"level":"ERROR","time":"2018-01-02T12:34:56.789+0800","msg":"errorf 5","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"PANIC","time":"2018-01-02T12:34:56.789+0800","msg":"panicf 6","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanicf"}
{"level":"ERROR","time":"2018-01-02T12:34:56.789+0800","msg":"error||5","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestDefaultLoggerWithJSON"}
{"level":"PANIC","time":"2018-01-02T12:34:56.789+0800","msg":"panic||6","file":"defaultlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.callPanic"}
`; expected != actual2 {
		t.Fatalf("invalid buf2.\nexpected:\n%v\nactual:\n%v", expected, actual2)
	}
}

func call(c chan bool, f func()) {
	defer func() {
		if e := recover(); e != nil {
			c <- true
		} else {
			c <- false
		}
	}()
	f()
	c <- false
}

func callPanicf(l *defaultLogger) {
	l.Panicf("panicf %v", 6)
}

func callPanic(l *defaultLogger) {
	l.Panic("panic", 6)
}

const rotateLogPath = "./log/rotate.log"
const rotateMovedLogPath = "./log/rotate_moved.log"

func TestDefaultLoggerRotate(t *testing.T) {
	var content1, content2 string

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	os.Remove(rotateLogPath)
	os.Remove(rotateMovedLogPath)

	swapCreatedLoggers(func() {
		Init(&Config{
			FilePath: rotateLogPath,
			Debug:    true,
		})

		rotateLogger := New(rotateLogPath)
		content1 = fmt.Sprint(rnd.Int63())
		rotateLogger.Print(content1)

		// 强制 flush 一下日志。
		time.Sleep(10 * time.Millisecond)
		globalConfigurer.(*defaultLoggerConfigurer).sync()

		if err := os.Rename(rotateLogPath, rotateMovedLogPath); err != nil {
			t.Fatalf("fail to rename log. [err:%v]", err)
		}

		// 故意创建一个空文件，用来测试 odin 日志切分的默认模式。
		f, err := os.OpenFile(rotateLogPath, os.O_CREATE, 0644)

		if err != nil {
			t.Fatalf("fail to create empty file. [err:%v]", err)
		}

		f.Close()
		time.Sleep(defaultLoggerMinRotateInterval + 500*time.Millisecond)

		content2 = fmt.Sprint(rnd.Int63())
		rotateLogger.Print(content2)
	})

	if content, err := ioutil.ReadFile(rotateLogPath); err != nil {
		t.Fatalf("fail to read file. [err:%v] [path:%v]", err, rotateLogPath)
	} else if actual := string(bytes.TrimSpace(content)); actual != content2 {
		t.Fatalf("unexpected content in file. [expected:%v] [actual:%v]", content2, actual)
	}

	if content, err := ioutil.ReadFile(rotateMovedLogPath); err != nil {
		t.Fatalf("fail to read file. [err:%v] [path:%v]", err, rotateMovedLogPath)
	} else if actual := string(bytes.TrimSpace(content)); actual != content1 {
		t.Fatalf("unexpected content in file. [expected:%v] [actual:%v]", content1, actual)
	}
}
