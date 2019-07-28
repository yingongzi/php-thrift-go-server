package log

import (
	"context"
	"io"
	"strings"
	"testing"

	"git.xiaojukeji.com/soda-framework/go-trace"
)

func TestPublicLog(t *testing.T) {
	var bufs []*closableBuffer

	swapCreatedLoggers(func() {
		Use(&globalConfigure{
			creator: func(string) io.WriteCloser {
				buf := &closableBuffer{}
				bufs = append(bufs, buf)
				return buf
			},
		})
		publicLogger = New("./log/public.log").(*delegateLogger)
		publicLogTestTimestamp = "2018-01-02 12:34:56"
		defer func() {
			publicLogTestTimestamp = ""
		}()

		ctx := context.Background()
		ctx = trace.WithInfo(ctx, &trace.Info{
			Traceid: "abcdef",
			Spanid:  "0987654321",
			Logid:   "1234567890",
		})

		tag := "_public_log_test"
		kvs := map[string]interface{}{
			"foo": 123456,
		}
		Public(ctx, tag, kvs)
		Public(ctx, tag, nil)
		Public(trace.WithInfo(ctx, &trace.Info{HintCode: 3}), tag, kvs)
	})

	lines := make([]string, 0, len(bufs))

	for i, buf := range bufs {
		lines = append(lines, buf.String())

		if !buf.closed {
			t.Fatalf("buf is not closed. [idx:%v]", i)
		}
	}

	if expected, actual := `_public_log_test_shadow||timestamp=2018-01-02 12:34:56||traceid=abcdef||spanid=0987654321||hintCode=3||foo=123456
_public_log_test||timestamp=2018-01-02 12:34:56||traceid=abcdef||spanid=0987654321
_public_log_test||timestamp=2018-01-02 12:34:56||traceid=abcdef||spanid=0987654321||foo=123456`, sortLines(maskDateAndFileLine(strings.Join(lines, "\n"))); expected != actual {
		t.Fatalf("invalid log output.\nexpected:\n%v\nactual:\n%v", expected, actual)
	}
}

func TestFormatFields(t *testing.T) {
	ctx := context.Background()
	ctx = trace.WithInfo(ctx, &trace.Info{
		Traceid: "abcdef",
		Spanid:  "0987654321",
		Logid:   "1234567890",
	})
	tag := "_format_fields_test"

	if actual, expected := FormatFields(ctx, tag, nil), tag+`||traceid=abcdef||spanid=0987654321`; actual != expected {
		t.Fatalf("invalid result. [expected:%v] [actual:%v]", expected, actual)
	}

	if actual, expected := FormatFields(ctx, tag, map[string]interface{}{
		"foo": 123456,
	}), tag+`||traceid=abcdef||spanid=0987654321||foo=123456`; actual != expected {
		t.Fatalf("invalid result. [expected:%v] [actual:%v]", expected, actual)
	}
}
