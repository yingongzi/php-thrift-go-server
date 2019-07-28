package trace

import (
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	processID     int
	traceSequence int32
	traceSource   byte = 0xda // soDA
)

func init() {
	processID = os.Getpid() & (1<<16 - 1)
}

// SetTraceidSource 设置 traceid 中的 source byte 并返回之前设置的值。
// 这个 source byte 用来标识一个业务线，默认是 0xda 代表 soda。
func SetTraceidSource(src byte) byte {
	old := traceSource
	traceSource = src
	return old
}

// Traceid 是一个符合公司规范的 traceid。
// 规则详见：http://wiki.intra.xiaojukeji.com/pages/viewpage.action?pageId=91930217
type Traceid string

// MakeTraceid 生成一个新的 traceid。
func MakeTraceid(unixnano int64) Traceid {
	ip, _ := GuessIP()
	ts := (unixnano / int64(time.Second)) & (1<<32 - 1)
	reserved := 0
	seq := atomic.AddInt32(&traceSequence, 1) & (1<<24 - 1)

	buf := make([]byte, 32)
	b := buf
	b = format(b, uint64(ip[0]), 1)
	b = format(b, uint64(ip[1]), 1)
	b = format(b, uint64(ip[2]), 1)
	b = format(b, uint64(ip[3]), 1)
	b = format(b, uint64(ts), 4)
	b = format(b, uint64(reserved), 2)
	b = format(b, uint64(processID), 2)
	b = format(b, uint64(seq), 3)
	b = format(b, uint64(traceSource), 1)
	tid := makeString(buf)

	return Traceid(tid)
}

// String 返回 traceid 的字符串值。
func (tid Traceid) String() string {
	return string(tid)
}

// IsValid 判断 traceid 是否合法。
// 这里只做最基本的长度判断，其他的不做验证。
func (tid Traceid) IsValid() bool {
	return len(tid) == 32
}

// IP 从 traceid 中解析出 IP 地址，如果无法解析则返回 nil。
func (tid Traceid) IP() net.IP {
	if !tid.IsValid() {
		return nil
	}

	b1, err1 := strconv.ParseInt(string(tid[:2]), 16, 8)
	b2, err2 := strconv.ParseInt(string(tid[2:4]), 16, 8)
	b3, err3 := strconv.ParseInt(string(tid[4:6]), 16, 8)
	b4, err4 := strconv.ParseInt(string(tid[6:8]), 16, 8)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil
	}

	return net.IPv4(byte(b1), byte(b2), byte(b3), byte(b4))
}

// Timestamp 解析 traceid 中的时间戳，如果无法解析则返回 0。
func (tid Traceid) Timestamp() int {
	if !tid.IsValid() {
		return 0
	}

	ts, err := strconv.ParseInt(string(tid[8:16]), 16, 32)

	if err != nil {
		return 0
	}

	return int(ts)
}

// ProcessID 返回一个 [0, 65535] 取值范围的值，这个值来源于 PID，但是不一定总是相等。
func (tid Traceid) ProcessID() int {
	if !tid.IsValid() {
		return 0
	}

	pid, err := strconv.ParseInt(string(tid[20:24]), 16, 32)

	if err != nil {
		return 0
	}

	return int(pid)
}

// SequenceID 返回 traceid 的序列号。
func (tid Traceid) SequenceID() int {
	if !tid.IsValid() {
		return 0
	}

	seq, err := strconv.ParseInt(string(tid[24:30]), 16, 32)

	if err != nil {
		return 0
	}

	return int(seq)
}

// Source 返回 traceid 的 source 标识。
func (tid Traceid) Source() byte {
	if !tid.IsValid() {
		return 0
	}

	src, err := strconv.ParseUint(string(tid[30:32]), 16, 8)

	if err != nil {
		return 0
	}

	return byte(src)
}
