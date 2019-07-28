package trace

import (
	"strconv"
)

const (
	ultronHintCodeMask          = 1
	emptyHintCode      HintCode = 0
)

// HintCode 表示一个 hint code, 是一个 64bit 整形，在trace链路中是透传的，生成之后不会再变化。
type HintCode int64

// MakeHintCode 生成一个hintCode, 用作流量标识。
func MakeHintCode(hintCode string) HintCode {
	if hintCode == "" {
		return emptyHintCode
	}

	hc, err := strconv.ParseInt(hintCode, 10, 64)

	if err != nil {
		return emptyHintCode
	}

	return HintCode(hc)
}

// MakeUltronHintCode 生成一个合法的 ultron hint code。
func MakeUltronHintCode() HintCode {
	return ultronHintCodeMask
}

// IsEmpty 判断这个 hintCode 是否是设置，如果返回 true 代表未设置。
func (hintCode HintCode) IsEmpty() bool {
	return hintCode == emptyHintCode
}

// IsUltron 判断这个 hintCode 是否是来自 ultron 的压测流量。
// 当 hintCode 最低位不为 0 时代表是 ultron 压测流量。
func (hintCode HintCode) IsUltron() bool {
	return hintCode&ultronHintCodeMask != 0
}

// String 返回 hintCode 的字符串值。
func (hintCode HintCode) String() string {
	return strconv.FormatInt(int64(hintCode), 10)
}
