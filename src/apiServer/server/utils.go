package UserCenter

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/astaxie/beego/orm"

	"modle/session"
	"modle/utils"
	"misc/mylog"
	"server/config"
	"server/gdb"
)

type RETCODE int

const (
	RET_CODE_Ok            RETCODE = 0   //成功
	RET_CODE_ServerErr     RETCODE = 1   //服务器自身错误
	RET_CODE_UnknowCommand RETCODE = 2   //未知的命令
	RET_CODE_ParseJsonErr  RETCODE = 3   //解析参数json失败
	RET_CODE_TokenInvalid  RETCODE = 4   //内部token无效，无此token或者token过期，需要重新登陆
	RET_CODE_ParamError    RETCODE = 5   //解析请求参数失败，可能是参数缺失或者参数格式/类型错误
	RET_CODE_TodoError     RETCODE = 10  //未完成的功能
	RET_CODE_3rdAuthFaile  RETCODE = 101 //因为第三方认证失败导致的登陆失败
	RET_CODE_CantFindUser  RETCODE = 102 //找不到指定的用户
)

type TokenData struct {
	Token     string
	Acc       *gdb.DBAccount
	User      *gdb.DBUser
	SessStore session.Store
}

func createAccAndUser(accName, password, string, req *http.Request) (acc *gdb.DBAccount, user *gdb.DBUser, err error) {
	ip := getIpFromRequest(req)

	if acc, err = gdb.NewAccount(accName, password); err != nil {
		return nil, nil, err
	}

	if user, err = gdb.NewUser(acc.AccId, ip); err != nil {
		return nil, nil, err
	}

	return acc, user, err
}

func getIpFromRequest(req *http.Request) (ip string) {
	ip = req.Header.Get("Remote_addr")
	if ip == "" {
		ip = req.RemoteAddr
	}
	return ip
}

var idMapToken map[int64]string = map[int64]string{}
var globalSessions *session.Manager

func initSession(provideName string, cf *session.ManagerConfig, a_onTimeOutCallBack session.CallBackFunc) {
	gob.Register(&TokenData{})
	gob.Register(&gdb.DBAccount{})
	gob.Register(&gdb.DBUser{})
	gob.Register(&qtree.Point{})

	globalSessions, _ = session.NewManager(provideName, cf, a_onTimeOutCallBack)
	go globalSessions.GC()
}

func getSession(token string) (*TokenData, bool) {
	store, err := globalSessions.GetSessionStore(token)
	if err != nil {
		return nil, false
	}

	tokenData := &TokenData{Token: token, SessStore: store}

	if v := tokenData.SessStore.Get("acc"); v != nil {
		tokenData.Acc = v.(*gdb.DBAccount)
	}
	if v := tokenData.SessStore.Get("user"); v != nil {
		tokenData.User = v.(*gdb.DBUser)
	}
	if v := tokenData.SessStore.Get("sceneNode"); v != nil {
		tokenData.SceneNode = v.(*qtree.Point)
	}

	return tokenData, true
}

func updateSession(tokenData *TokenData) {
	if tokenData == nil {
		return
	}
	if tokenData.SessStore == nil {
		return
	}

	mylog.Debug("updateSession:", tokenData.Token)
	tokenData.SessStore.Set("acc", tokenData.Acc)
	tokenData.SessStore.Set("user", tokenData.User)
	globalSessions.SessionUpdate(tokenData.SessStore)
}

func onLogin(tokenData *TokenData) {
	store, err := globalSessions.SessionCreate()
	if err != nil {
		fmt.Printf("get session=%v\n", err)
		return
	}
	tokenData.SessStore = store
	tokenData.Token = tokenData.SessStore.SessionID()
	tokenData.SceneNode = qtree.NewPoint(tokenData.Acc.LastLongitude, tokenData.Acc.LastLatitude, tokenData.Token)
	Scene.Insert(tokenData.SceneNode)
	tokenData.Acc.UpdateInfo("LastToken", tokenData.Token, false)

	if config.DEBUG_FLAG {
		fmt.Printf("....111 onLogin,user=%v\n", tokenData.User.UserId)
	}
}

func onLogout(tokenData *TokenData) {
	if tokenData.SceneNode != nil {
		Scene.Remove(tokenData.SceneNode)
	}
	if tokenData.Acc != nil {
		tokenData.Acc.UpdateInfo("LastToken", "", false)
	}
	if tokenData.SessStore != nil {
		globalSessions.SessionDestroy(tokenData.SessStore)
	}
}
