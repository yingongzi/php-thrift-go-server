package trace

// Logid 表示一个 log id，当前的实现规则是将 time.Now().UnixNano() 的十六进制字符串值作为 log id。
type Logid string

// MakeLogid 根据 unixnano 生成一个新的 log id。
func MakeLogid(unixnano int64) Logid {
	return Logid(hexString(uint64(unixnano)))
}

// String 返回 log id 的字符串值。
func (logid Logid) String() string {
	return string(logid)
}
