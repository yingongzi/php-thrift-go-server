package log

import (
	"testing"
)

func ExampleLogger() {
	// myLogger 可以是一个全局变量。
	myLogger := New("./log/mylogger.log")

	// 使用这个 logger。
	myLogger.Debug("k1=v1", "k2=v2")

	// 应该在 main 里面初始化 logger。
	Init(&Config{
		FilePath:     "./log/all.log",
		ShowFileLine: true,
		Debug:        true,
	})

	Debugf("default log in ./log/all.log")
	Errorf("error log in both ./log/all.log and ./log/error.log")
	Printf("print some%v", "thing")
	myLogger.Infof("my log||foo=%v", "bar")
	myLogger.Errorf("my another error log||foo=%v", "bar")
	myLogger.Print("some", "log")

	Close()

	// Output:
}

func TestLogNew(t *testing.T) {
	swapCreatedLoggers(func() {
		New("")
		New("./log/foo.log")
		New("./log/all.log")
		New("./log/all.log")
		New("./log/foo.log")

		Init(&Config{
			FilePath:     "./log/all.log",
			ShowFileLine: true,
			Debug:        true,
		})

		if expected := 3; len(createdLoggers) != expected {
			t.Fatalf("invalid count of created loggers. [expected:%v] [actual:%v]", expected, createdLoggers)
		}

		New("./log/foo.log")
		New("./log/bar.log")

		if expected := 4; len(createdLoggers) != expected {
			t.Fatalf("invalid count of created loggers. [expected:%v] [actual:%v]", expected, createdLoggers)
		}
	})
}
