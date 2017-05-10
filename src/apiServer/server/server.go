package apiServer

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"net/http"

	"misc/mylog"
	"misc/json"
	"modle/session"
	_ "modle/session/redis"
	"modle/utils"
	"server/config"
	"server/gdb"
)

type procCallBack func(req *http.Request, tokenData *TokenData, data interface{}) (code RETCODE, retsult interface{}, err error)

type respData struct {
	proc              procCallBack
	needCheckToken    bool
	needUpdateSession bool
	dataType          reflect.Type
}

var routerMap map[string]*respData = map[string]*respData{}

type TReq struct {
	Param interface{} `json:"param"` //Param *json.RawMessage
	Token string      `json:"token"`
	// Sign	string 			`json:"sign"`
}

type TRet struct {
	Code   RETCODE     `json:"code"`
	Data   interface{} `json:"data"`
	ErrMsg string      `json:"errMsg"`
}

func mainProc(w http.ResponseWriter, req *http.Request) {
	var needEncode bool = true
	var ret TRet
	defer func() {
		if r := recover(); r != nil {
			mylog.Error("mainProc: recover error:", r)
			return
		}

		if !needEncode {
			return
		}

		var err error
		var retByte []byte
		if retByte, err = json.Marshal(ret, false); err != nil {
			io.WriteString(w, fmt.Sprintf("{\"code\": %v,\"msg\": \"failed\"}", RET_CODE_ServerErr))
			mylog.Error("json marshal ret error:", err.Error())
		} else {
			if _, err := w.Write(retByte); err != nil {
				mylog.Error("write ret byte error:", err.Error())
			}
		}
		mylog.Debug("-------------------------------------- end ------------------------------------")
	}()

	mylog.Debug("---------------------------------- rev request --------------------------------")
	mylog.Debug("cmd=", req.URL.Path)

	respData, ok := routerMap[req.URL.Path]
	if !ok {
		ret.Code = RET_CODE_UnknowCommand
		ret.ErrMsg = string(req.URL.Path)
		return
	}

	reqByte, err := ioutil.ReadAll(req.Body)
	if err != nil {
		needEncode = false
		io.WriteString(w, "{\"code\": RET_CODE_ServerErr,\"msg\": \"io read err\"}")
		return
	}

	mylog.Debug("string(reqByte)=", string(reqByte))
	var reqData TReq
	if respData.dataType != nil {
		reqData.Param = reflect.New(respData.dataType).Interface()
	}
	if err = json.Unmarshal(reqByte, &reqData); err != nil {
		ret.Code = RET_CODE_ParseJsonErr
		return
	}

	var tokenData *TokenData
	if respData.needCheckToken {
		var exist bool
		if tokenData, exist = getSession(reqData.Token); !exist || tokenData == nil {
			ret.Code = RET_CODE_TokenInvalid
			return
		}
	}

	mylog.Debug("handle start")

	ret.Code, ret.Data, err = respData.proc(req, tokenData, reqData.Param)
	mylog.Debug("handle complate, Code, Data, err=", ret.Code, ret.Data, err)
	if ret.Code != RET_CODE_Ok {
		if err != nil {
			ret.ErrMsg = err.Error()
		}
	} else {
		ret.ErrMsg = "ok"
	}
	if respData.needUpdateSession {
		updateSession(tokenData)
	}

}

func Start(addr string) (err error) {
	Init()

	cfg := config.C_MysqlCfg
	utils.OpenServerDB(cfg.MysqlAddr, cfg.MaxIdleConns, cfg.MaxOpenConns, cfg.Debug, false, gdb.Register)


	providerConfig := fmt.Sprintf("%v,%v,%v,%v", config.C_RedisCfg.Address, config.C_RedisCfg.MaxPoolSize, config.C_RedisCfg.Password, config.C_RedisCfg.Dbnum)
	initSession("redis",
		&session.ManagerConfig{Maxlifetime: config.C_ExCfg.SessionTimeOut, Gclifetime: config.C_ExCfg.SessionTimeOut / 10, ProviderConfig: providerConfig},
		onSessionTimeOut)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		mylog.Error("Start server err:", err.Error)
		log.Fatal("ListenAndServe: ", err)
		return err
	}

	return nil
}

func onSessionTimeOut(key string, tokenDataInterface interface{}) {
	if tokenData, ok := getSession(key); tokenData != nil && ok {
		onLogout(tokenData)
	}
}

func registerHandle(router string, needCheckToken_ bool, needUpdateSession_ bool, dataTypeTemp interface{}, proc_ procCallBack) (err error) {
	if _, exist := routerMap[router]; exist {
		fmt.Println("re register:" + router)
		mylog.Error("re register:" + router)
		return fmt.Errorf("re register:" + router)
	}
	http.HandleFunc(router, mainProc)

	routerMap[router] = &respData{
		proc:              proc_,
		needCheckToken:    needCheckToken_,
		needUpdateSession: needUpdateSession_,
		dataType:          reflect.TypeOf(dataTypeTemp),
	}

	return nil
}

func Init() {
	registerHandle("/acc/auth", false, false, tReqAuthType{}, onAuthHandle)
	registerHandle("/acc/logout", true, false, nil, onAccLogoutHandle)
	registerHandle("/user/getInfo", true, true, tGetUserInfoType{}, onGetUserInfoHandle)
}

// curl -d '{"a":"aaa","b":0}' 127.0.0.1:7777/hello
// curl -H "Content-Type: application/json" -X POST  --data '{"data":"1"}' http://127.0.0.1:7777/auth
// curl -H "Content-Type: application/x-www-form-urlencoded" -X POST  --data '{\"data\":\"1\"}' http://127.0.0.1:7777/main
// curl -i -X POST -H "'Content-type':'application/x-www-form-urlencoded', 'charset':'utf-8', 'Accept': 'text/plain'" -d 'json_data={"a":"aaa","b":1}' url
//
// curl -i -X POST -H "'Content-type':'application/x-www-form-urlencoded', 'Accept': 'text/plain'" -d json_data={"a":"aaa","b":1} http://127.0.0.1:7777/main

// curl -H "Content-Type: application/x-www-form-urlencoded" -X POST  --data '{\"cmd\":\"auth\", \"channel\":\"qq\", \"threeToken\":\"asdfawsf4e\"}' http://127.0.0.1:7777/main
// curl -d {\"cmd\":\"auth\",\"channel\":\"qq\",\"threeToken\":\"asdfawsf4e\"} http://127.0.0.1:7777/main
