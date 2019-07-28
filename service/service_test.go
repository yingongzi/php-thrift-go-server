package service

import (
	"fmt"
	"git.xiaojukeji.com/soda-framework/go-log"
	"github.com/yingongzi/php-thrift/gen-go/php_go/idl"
	"php-thrift-go-server/client"
	"php-thrift-go-server/conf"
	"php-thrift-go-server/util"
	"testing"
)

func TestService_SetUsers(t *testing.T) {
	//加载配置文件
	conf.LoadConfigFile("../conf/service.conf")
	config := conf.GoServerConf
	//log 模块的初始化
	log.Init(&config.LogConf)
	defer log.Close()
	//Redis模块的初始化
	client.InitRedis(config.RedisConf)

	svr := New()

	info01 := idl.UserInfo{
		UserID:   1,
		Username: "sunyin01",
		Age:      18,
		Gender:   false,
	}
	info02 := idl.UserInfo{
		UserID:   2,
		Username: "sunyin02",
		Age:      28,
		Gender:   true,
	}
	info03 := idl.UserInfo{
		UserID:   3,
		Username: "sunyin03",
		Age:      38,
		Gender:   true,
	}
	users := []idl.UserInfo{}
	users = append(users, info01, info02, info03)
	str := util.JsonString(users)
	resp, err := svr.SetUsers(&idl.SetUsersReq{UserInfoStr: str})
	fmt.Println(resp, "=======", err)

	resp2, err := svr.GetUserByUserID(&idl.GetUserByIdReq{UserID: 1})
	fmt.Println(util.JsonString(resp2), "=======", err)

}
