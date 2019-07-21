package conf

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"path/filepath"
	"php-thrift-go-server/util"
)

var (
	GoServerConf Config
)

type RedisConf struct {
	Addr string 	`toml:"addr"`
}

type LogConf struct {
	FilePath		 string 	`toml:"file_path"`
	ErrorFilePath	 string		`toml:"error_file_path"`
	Level			 string		`toml:"level"`
	MaxSizeMB		 int64		`toml:"max_size_mb"`
}


type Config struct {
	redisConf 	RedisConf		`toml:"redis_conf"`
	logConf 	LogConf			`toml:"log_conf"`
}

func LoadConfigFile() (err error) {
	confFilePath := "service.conf"
	absConfPath, err := filepath.Abs(confFilePath)
	fmt.Println("load config from ", absConfPath)


	tomlTree, err := toml.LoadFile(confFilePath)
	GoServerConf = Config{
		redisConf:RedisConf{
			Addr:tomlTree.Get("redis_conf.addr").(string),
		},
		logConf:LogConf{
			FilePath:tomlTree.Get("log_conf.file_path").(string),
			ErrorFilePath:tomlTree.Get("log_conf.error_file_path").(string),
			Level:tomlTree.Get("log_conf.level").(string),
			MaxSizeMB:tomlTree.Get("log_conf.max_size_mb").(int64),
		},
	}
	//node := tomlTree.Get("redis_conf.addr").(string)
	fmt.Println(util.JsonString(GoServerConf))
	return err
}