package main

import (
	"crypto/tls"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"git.xiaojukeji.com/soda-framework/go-log"
	"github.com/yingongzi/php-thrift/gen-go/php_go/idl"
	"php-thrift-go-server/client"
	"php-thrift-go-server/conf"
	"php-thrift-go-server/service"
)

const (
	ADDR = "localhost:8999"

)

func main()  {
	//加载配置文件
	conf.LoadConfigFile("conf/service.conf")
	config := conf.GoServerConf
	//log 模块的初始化
	log.Init(&config.LogConf)
	defer log.Close()
	//Redis模块的初始化
	client.InitRedis(config.RedisConf)

	// thrift 服务启动
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTTransportFactory()
	secure := false
	if err := runServer(transportFactory, protocolFactory, ADDR, secure); err != nil {
		fmt.Println("error running server:", err)
	}
}

func runServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool) error {
	var transport thrift.TServerTransport
	var err error
	if secure {
		cfg := new(tls.Config)
		if cert, err := tls.LoadX509KeyPair("server.crt", "server.key"); err == nil {
			cfg.Certificates = append(cfg.Certificates, cert)
		} else {
			return err
		}
		transport, err = thrift.NewTSSLServerSocket(addr, cfg)
	} else {
		transport, err = thrift.NewTServerSocket(addr)
	}

	if err != nil {
		return err
	}
	//fmt.Printf("%T\n", transport)

	handler := service.New()
	processor := idl.NewPhp_Go_SvrProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	fmt.Println("Starting the simple server... on ", addr)
	return server.Serve()
}
