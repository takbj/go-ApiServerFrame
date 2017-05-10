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


// func init() {
// 	//	fmt.Println("===config.init===")
// 	if err := ReloadUserCfg(); err != nil {
// 		os.Exit(-1)
// 	}
// 	registerCfg("User", "config/UserCfg.json", &UserCfg{})
// }

// func ReloadUserCfg() error {
// 	tmpCfg := UserCfg{}
// 	err := ReloadCfg("config/UserCfg.json", &tmpCfg)
// 	if err != nil {
// 		mylog.Error("server start error:ExCfg json file read failed,", err)
// 		fmt.Println("server start error:ExCfg json file read failed,", err)
// 	}else{
// 		C_ExCfg = tmpCfg
// 	}

// 	return err
// }
