package trace

import (
	"unsafe"
)

const (
	hexByteBits = 4
	uint64Bytes = 8
)

var (
	hexBytes = [1 << hexByteBits]byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f',
	}
)

// format 实现了一个远比 fmt.Sprintf 高效的字符串拼接算法，仅支持将 value 变成 16 进制的形式。
// 这里假定了 endianess 是 BigEndian，由于我们已经接触不到 LittleEndian 的设备，所以就没必要去支持了。
// 这里为了性能并不做任何的合法性检查，如果写越界了就直接 panic。
func format(dst []byte, value uint64, bytes int) []byte {
	data := (*[uint64Bytes]byte)(unsafe.Pointer(&value))[:bytes]

	for i := 0; i < bytes; i++ {
		v := data[bytes-i-1]
		dst[2*i] = hexBytes[v>>hexByteBits]
		dst[2*i+1] = hexBytes[v&(1<<hexByteBits-1)]
	}

	return dst[bytes*2:]
}

func makeString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

func hexString(value uint64) string {
	buf := make([]byte, 16)
	format(buf, value, 8)
	return makeString(buf)
}
