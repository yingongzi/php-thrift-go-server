package conf

import (
	"fmt"
	"git.xiaojukeji.com/soda-framework/go-log"
	"github.com/pelletier/go-toml"
	"php-thrift-go-server/util"
)

var (
	GoServerConf Config
)

type RedisConf struct {
	Addr string 	`toml:"addr"`
}

//type LogConf struct {
//	FilePath		 string 	`toml:"file_path"`
//	ErrorFilePath	 string		`toml:"error_file_path"`
//	Level			 string		`toml:"level"`
//	MaxSizeMB		 int64		`toml:"max_size_mb"`
//}


type Config struct {
	RedisConf 	RedisConf		`toml:"redis_conf"`
	LogConf 	log.Config		`toml:"log_conf"`
}

func LoadConfigFile(path string) (err error) {
	//confFilePath := "conf/service.conf"
	//absConfPath, err := filepath.Abs(confFilePath)
	//fmt.Println("load config from ", absConfPath)


	tomlTree, err := toml.LoadFile(path)
	GoServerConf = Config{
		RedisConf:RedisConf{
			Addr:tomlTree.Get("redis_conf.addr").(string),
		},
		LogConf:log.Config{
			FilePath:tomlTree.Get("log_conf.file_path").(string),
			ErrorFilePath:tomlTree.Get("log_conf.error_file_path").(string),
			Level:tomlTree.Get("log_conf.level").(string),
			MaxSizeMB:int(tomlTree.Get("log_conf.max_size_mb").(int64)),
		},
	}
	fmt.Println(util.JsonString(GoServerConf))
	//todo 此处我要打印这个config结果失败了，json解析不出来，不知道为啥
	return err
}