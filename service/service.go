package service

import (
	"git.xiaojukeji.com/soda-framework/go-log"
	"github.com/yingongzi/php-thrift/gen-go/php_go/idl"
	"php-thrift-go-server/rpc"
	"php-thrift-go-server/util"
	"strconv"
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

func(*Service) GetUserByUserID(req *idl.GetUserByIdReq)(resp *idl.GetUserByIdResp, err error){
	log.Infof("Service||GetUserByUserID||req=%v", util.JsonString(req))
	resp = &idl.GetUserByIdResp{
		Header:&idl.ResponseHeader{},
		User:&idl.UserInfo{},
	}
	val, err := rpc.RedisGet(strconv.FormatInt(int64(req.UserID), 10))
	if err != nil {
		resp.Header.Code = 1
		resp.Header.Msg = "get value from redis error"
		log.Errorf("Service||GetUserByUserID||redis internal error||userID=%d", req.UserID)
		return
	}
	user := idl.UserInfo{}
	err = util.JsonUnmarshalFromString(val, &user)
	if err != nil {
		resp.Header.Code = 2
		resp.Header.Msg = "util.JsonUnmarshalFromString error"
		log.Errorf("Service||GetUserByUserID||util.JsonUnmarshalFromString error||user=%d", val)
		return
	}
	resp.Header.Code = 0
	resp.User = &user
	return
}

func(*Service) SetUsers(req *idl.SetUsersReq)(resp *idl.SetUsersResp, err error){
	log.Infof("Service||SetUsers||req=%v", util.JsonString(req))
	resp = &idl.SetUsersResp{
		Header:&idl.ResponseHeader{},
		UserIDs:[]int32{},
	}
	users := []idl.UserInfo{}
	err = util.JsonUnmarshalFromString(req.UserInfoStr, &users)
	if err != nil {
		resp.Header.Code = 3
		resp.Header.Msg = "JsonUnmarshalFromString error"
		log.Errorf("Service||SetUsers||util.JsonUnmarshalFromString error||users=%v", req.UserInfoStr)
		return
	}
	for _, user := range users {
		rpc.RedisSet(strconv.FormatInt(int64(user.UserID), 10), user)
		resp.UserIDs = append(resp.UserIDs, user.UserID)
	}
	resp.Header.Code = 0
	return
}
