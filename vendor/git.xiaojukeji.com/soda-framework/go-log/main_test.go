package log

import (
	"bytes"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type closableBuffer struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	closed bool
}

func (buf *closableBuffer) Write(data []byte) (int, error) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	if buf.closed {
		return 0, errors.New("closed")
	}

	return buf.buf.Write(data)
}

func (buf *closableBuffer) Close() error {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.closed = true
	return nil
}

func (buf *closableBuffer) String() string {
	return buf.buf.String()
}

func sortLines(s string) string {
	lines := strings.Split(s, "\n")
	sort.Strings(lines)

	for lines[0] == "" {
		lines = lines[1:]
	}

	return strings.Join(lines, "\n")
}

func swapCreatedLoggers(fn func()) {
	oldLoggers := createdLoggers
	createdLoggers = map[string]*delegateLogger{}
	defer func() {
		Close()

		for _, logger := range createdLoggers {
			logger.Close()
		}

		createdLoggers = oldLoggers
	}()

	fn()
}

var (
	reMaskDate  = regexp.MustCompile(`\[\d+[^\]]+\]`)
	reMaskLines = regexp.MustCompile(`:\d+@`)
)

func maskDateAndFileLine(s string) string {
	return reMaskDate.ReplaceAllString(reMaskLines.ReplaceAllString(s, ":0@"), "[]")
}
