package trace

// HintContent 用作通用数据透传通路，是序列化的json字串，请求源头或者链路上任意模块都可以添加修改hintContent
type HintContent string

// MakeHintContent 生成hintContent
func MakeHintContent(hintContent string) HintContent {
	return HintContent(hintContent)
}

// String 返回 hintContent 的字符串值。
func (hintContent HintContent) String() string {
	return string(hintContent)
}
