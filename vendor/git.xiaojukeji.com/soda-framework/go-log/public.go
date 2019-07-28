package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"git.xiaojukeji.com/soda-framework/go-trace"
)

var (
	publicLogger = New("./log/public.log").(*delegateLogger)

	publicSeparator       = []byte("||")
	publicKeyTimestamp    = []byte("timestamp=")
	publicLogUltronSuffix = []byte("_shadow")
)

const (
	publicTimeFormat           = "2006-01-02 15:04:05"
	publicLogFormatArgsLogSize = 1 << 12
)

var (
	// 用来测试的时间戳，如果设置了就会使用这个值而不是 time.Now()。
	publicLogTestTimestamp = ""
)

// Public 输出 public.log。
func Public(ctx context.Context, tag string, kvs map[string]interface{}) {
	buf := &bytes.Buffer{}
	buf.Grow(publicLogFormatArgsLogSize)

	// 日志格式：
	// ${tag}||timestamp=${now}||${ctx}||${for-each kvs}
	formatTag(ctx, buf, tag)
	buf.Write(publicSeparator)
	buf.Write(publicKeyTimestamp)

	if publicLogTestTimestamp == "" {
		buf.WriteString(time.Now().Format(publicTimeFormat))
	} else {
		buf.WriteString(publicLogTestTimestamp)
	}

	buf.Write(publicSeparator)
	buf.WriteString(trace.ContextString(ctx))
	formatKVMap(buf, kvs)

	if writer, ok := publicLogger.logger.(io.Writer); ok {
		writer.Write(buf.Bytes())
	} else {
		publicLogger.logger.Print(buf.String())
	}
}

// FormatFields 会生成符合公司日志规范格式的日志。
func FormatFields(ctx context.Context, tag string, kvs map[string]interface{}) string {
	buf := &bytes.Buffer{}
	buf.Grow(publicLogFormatArgsLogSize)
	formatTag(ctx, buf, tag)

	buf.Write(publicSeparator)
	buf.WriteString(trace.ContextString(ctx))
	formatKVMap(buf, kvs)

	return buf.String()
}

func formatTag(ctx context.Context, buf *bytes.Buffer, tag string) {
	buf.WriteString(tag)

	if trace.ContextHintCode(ctx).IsUltron() {
		buf.Write(publicLogUltronSuffix)
	}
}

func formatKVMap(buf *bytes.Buffer, kvs map[string]interface{}) {
	for k, v := range kvs {
		buf.Write(publicSeparator)
		buf.WriteString(k)
		buf.WriteRune('=')
		fmt.Fprint(buf, v)
	}
}
