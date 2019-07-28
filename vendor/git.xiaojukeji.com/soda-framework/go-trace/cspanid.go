package trace

// CSpanid 从上游传递过来，下游会将这个信息当做 spanid 使用。
type CSpanid string

// MakeCSpanid 根据 unixnano 来生成一个新的 cspanid。
func MakeCSpanid(unixnano int64) CSpanid {
	return CSpanid(MakeSpanid(unixnano))
}

// String 返回 cspanid 的字符串值。
func (cspanid CSpanid) String() string {
	return string(cspanid)
}

// IsValid 判断 cspanid 是否合法，当前只检查长度是否正确。
func (cspanid CSpanid) IsValid() bool {
	return len(cspanid) == 16
}
