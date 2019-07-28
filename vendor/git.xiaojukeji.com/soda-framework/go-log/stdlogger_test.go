package log

import (
	"io"
	"log"
	"strings"
	"testing"
	"time"
)

func TestStdLogger(t *testing.T) {
	var bufs []*closableBuffer

	swapCreatedLoggers(func() {
		Use(&globalConfigure{
			creator: func(string) io.WriteCloser {
				buf := &closableBuffer{}
				bufs = append(bufs, buf)
				return buf
			},
		})
		l := New("")
		logger := NewStdLogger(l, INFO, 0, log.Lshortfile)

		// calldepth 并不工作。
		logger.Output(0, "first line")
		logger.Output(1, "second line")
		logger.Output(1000, "third line")

		// 没有办法刷新 std logger，只能等一会。
		time.Sleep(10 * time.Millisecond)
	})

	lines := make([]string, 0, len(bufs))

	for i, buf := range bufs {
		lines = append(lines, buf.String())

		if !buf.closed {
			t.Fatalf("buf is not closed. [idx:%v]", i)
		}
	}

	if expected, actual := `[INFO][][stdlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestStdLogger.func1] first line
[INFO][][stdlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestStdLogger.func1] second line
[INFO][][stdlogger_test.go:0@git.xiaojukeji.com/soda-framework/go-log.TestStdLogger.func1] third line`, sortLines(maskDateAndFileLine(strings.Join(lines, "\n"))); expected != actual {
		t.Fatalf("invalid log output.\nexpected:\n%v\nactual:\n%v", expected, actual)
	}
}
