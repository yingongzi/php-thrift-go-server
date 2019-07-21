namespace go php_go.idl
namespace php php_go.idl

struct UserInfo
{
    1:i32 userID;
    2:string username;
    3:i32 age;
    4:bool gender;
}

struct ResponseHeader
{
    1:i32 code;
    2:string msg;
}

struct GetUserByIdReq{
    1: required i32    userID;    //用户id
}

struct GetUserByIdResp {
    1: required ResponseHeader header;
    2: required UserInfo user;
}

struct SetUsersReq{
    1: required string userInfoStr;
}
struct SetUsersResp{
    1: required ResponseHeader header;
    2: required list<i32> userIDs;
}


service Php_Go_Svr
{
    GetUserByIdResp GetUserByUserID(1:required GetUserByIdReq req)
    SetUsersResp SetUsers(1:required SetUsersReq req)
}