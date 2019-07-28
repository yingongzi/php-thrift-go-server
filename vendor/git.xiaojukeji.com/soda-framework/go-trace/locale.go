package trace

const (
	emptyLocale Locale = ""
)

// Locale 表示一个标准的区域信息。
//
// 当前我们不封装任何的解析方法，如果需要从 Locale 中解析出有意义的信息，
// 应该使用 golang.org/x/text/language 库，
// 详见 https://blog.golang.org/matchlang。
type Locale string

// MakeLocale 将一个字符串包装成 Locale。
// 这个函数不对 locale 的合法性做任何的校验。
func MakeLocale(locale string) Locale {
	return Locale(locale)
}

// IsEmpty 判断这个 locale 是否是设置，如果返回 true 代表未设置。
func (locale Locale) IsEmpty() bool {
	return locale == emptyLocale
}

// String 返回 l 的字符串。
func (l Locale) String() string {
	return string(l)
}
