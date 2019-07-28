package log

// Config 的各种默认配置。
const (
	DefaultFilePath      = "./log/all.log"
	DefaultErrorFilePath = "./log/error.log"
	DefaultLevel         = "debug"
	DefaultFormatter     = "text"

	defaultMaxSizeMB = 1024 * 1024 * 1024
)

// Config 是默认 Logger 的配置信息。
type Config struct {
	FilePath      string `toml:"file_path"`       // 日志文件名，默认为 ./log/all.log。
	ErrorFilePath string `toml:"error_file_path"` // 错误日志文件名，默认为 ./log/error.log。如果这个值与 FilePath 相同，那么会禁用独立的 error log。
	Level         string `toml:"level"`           // 日志级别，包括 debug、info、warning、error、fatal、panic，默认是 debug。
	MaxSizeMB     int    `toml:"max_size_mb"`     // 日志文件最大容量，默认是 0，不做最大日志自动切分。
	MaxBackups    int    `toml:"max_backups"`     // 最大日志备份文件个数，默认 0 个。
	Formatter     string `toml:"formatter"`       // 日志格式，可以是 text 和 json，默认为 text。
	ShowFileLine  bool   `toml:"show_file_line"`  // 是否在日志每一行里面显示行号，默认 false。
	Debug         bool   `toml:"debug"`           // 是否处于调试状态，默认为 false。如果设置为 true，日志会同时输出到 stderr 里。
}
