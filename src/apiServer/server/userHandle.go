package server

import (
	"net/http"

	"apiServer/gdb"
)

type tGetUserInfoType struct {
	UserId int64 `json:"UserId"` //用户ID
}

func onGetUserInfoHandle(req *http.Request, selfTD *TokenData, data interface{}) (code RETCODE, ret interface{}, err error) {
	reqData := data.(*tGetUserInfoType)
	user, err := gdb.FindUserById(reqData.UserId)
	if err != nil {
		return RET_CODE_ServerErr, nil, err
	}
	if user == nil {
		return RET_CODE_CantFindUser, nil, nil
	}
	return RET_CODE_Ok, user, nil
}