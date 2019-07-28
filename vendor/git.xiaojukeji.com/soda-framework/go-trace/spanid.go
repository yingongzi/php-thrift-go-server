package trace

import (
	"time"
)

// Spanid 是一个符合公司规范的 spanid。
// 规则详见：http://wiki.intra.xiaojukeji.com/pages/viewpage.action?pageId=35426943
type Spanid string

// MakeSpanid 根据 unixnano 来生成一个新的 spanid。
func MakeSpanid(unixnano int64) Spanid {
	traceRandMutex.Lock()
	rnd := traceRand.Int63()
	traceRandMutex.Unlock()
	return makeSpanid(unixnano, rnd)
}

func makeSpanid(unixnano, rnd int64) Spanid {
	ip, _ := GuessIP()
	ts := uint64((unixnano / int64(time.Second))) & (1<<32 - 1)

	var id uint64
	id = uint64(ip[0]) ^ ((ts >> (32 - 8)) & (1<<8 - 1))
	id = id<<8 | (uint64(ip[1]) ^ ((ts >> (32 - 16)) & (1<<8 - 1)))
	id = id<<8 | (uint64(ip[2]) ^ ((ts >> (32 - 24)) & (1<<8 - 1)))
	id = id<<8 | (uint64(ip[3]) ^ ((ts >> (32 - 32)) & (1<<8 - 1)))
	id = id<<32 | (uint64(rnd) & (1<<32 - 1))

	return Spanid(hexString(id))
}

// String 返回 spanid 的字符串值。
func (spanid Spanid) String() string {
	return string(spanid)
}

// IsValid 判断 spanid 是否合法，当前只检查长度是否正确。
func (spanid Spanid) IsValid() bool {
	return len(spanid) == 16
}
