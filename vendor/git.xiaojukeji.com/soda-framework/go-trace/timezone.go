package trace

const (
	emptyTimezone Timezone = ""
)

// Timezone 表示时区信息
//
// 当前我们使用的时区为字符串，例如"Asia/Shanghai"，
// 全球时区地址：https://0x9.me/x7gLi
type Timezone string

// MakeTimezone 将一个字符串包装成 Timezone。
// 这个函数不对 timezone 的合法性做任何的校验。
func MakeTimezone(timezone string) Timezone {
	return Timezone(timezone)
}

// IsEmpty 判断这个 locale 是否是设置，如果返回 true 代表未设置。
func (timezone Timezone) IsEmpty() bool {
	return timezone == emptyTimezone
}

// String 返回 t 的字符串。
func (t Timezone) String() string {
	return string(t)
}
