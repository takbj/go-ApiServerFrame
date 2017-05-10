package apiServer

import (
	"misc/mylog"
	"net/http"
	"server/gdb"
)

//响应client登陆请求
type tReqAuthType struct {
	UserName string
	Password string
}

func onAuthHandle(req *http.Request, tokenData *TokenData, data interface{}) (code RETCODE, ret interface{}, err error) {
	reqData := data.(*tReqAuthType)

	//获取帐号数据，没有就创建帐号及用户
	var exist bool
	tokenData = &TokenData{}
	if exist, tokenData.Acc, err = gdb.GetAccout(reqData.UserName, reqData.Password); !exist {
		tokenData.Acc, tokenData.User, err = createAccAndUser(channel, reqData.ThreeRdToken, openid, Area, Longitude, Latitude, req, reqData.ThreeRdExInfo)
	}
	if err != nil {
		mylog.Error(err)
		return RET_CODE_ServerErr, nil, nil
	}
	tokenData.Acc.Set3RdToken(reqData.ThreeRdToken)

	loginNotify := true
	if tokenData.Acc.LastToken != "" {
		if tmp, ok := getSession(tokenData.Acc.LastToken); tmp != nil && ok {
			tokenData = tmp
			loginNotify = false
		}
	}

	if tokenData.User == nil {
		//获取用户数据
		if tokenData.User, err = gdb.FindUserByAcc(tokenData.Acc.AccId); err != nil {
			mylog.Error("error:authHandle, db error:", err)
			return RET_CODE_ServerErr, nil, nil
		}
	}

	if loginNotify {
		onLogin(tokenData)
	}
	updateSession(tokenData)

	mylog.Debug("tokenData.Token 222:", tokenData.Token)
	return RET_CODE_Ok, map[string]interface{}{"Token": tokenData.Token, "accInfo": tokenData.Acc, "userInfo": tokenData.User}, nil
}

func onAccLogoutHandle(req *http.Request, tokenData *TokenData, data interface{}) (code RETCODE, ret interface{}, err error) {
	onLogout(tokenData)
	return RET_CODE_Ok, nil, nil
}
