package config

// import (
// 	"fmt"
// 	"misc/mylog"
// 	"os"
// )

var (
	C_ExCfg ExCfg
)

type ExCfg struct {
	ServerAddr 					string  `json:"ServerAddr"`//服务器监听地址与端口
	SessionTimeOut				int64 `json:"SessionTimeOut"`	// = 172800 //session过期时长,2tian
}

func init(){
	registerCfg("ex", "config/excfg.json",	&C_ExCfg)
}
