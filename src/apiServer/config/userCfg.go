package config

// import (
// 	"fmt"
// 	"golib/json"
// 	"io/ioutil"
// 	"misc/mylog"
// 	"os"
// )

var (
	C_UserCfg UserCfg
)

type UserCfg struct {
	MinNameLen int  `json:"MinNameLen"`
	MaxNameLen int  `json:"MaxNameLen"`
	Sex        int8 `json:"Sex"`
	Age        int  `json:"Age"`
}
func init(){
	registerCfg("user", "config/UserCfg.json",	&C_UserCfg)
}
