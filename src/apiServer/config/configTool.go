package config

import (
	"fmt"
	"golib/json"
	"io/ioutil"
	"misc/mylog"
	"reflect"
	"os"
)

type tCfgMap struct{
	fileName string
	cfgDataRef interface{}
}

var allCfg map[string]tCfgMap = map[string]tCfgMap{}

func registerCfg(a_cfgName, a_fileName string, a_cfgDataRef interface{}){
	allCfg[a_cfgName] = tCfgMap{
		fileName:a_fileName,
		cfgDataRef:a_cfgDataRef,
	}
}

func Init(){
	needExit := false
	for cfgName, _ := range allCfg {
		if err:=ReloadCfg(cfgName); err!=nil {
			needExit = true
		}
	}

	if needExit {
		os.Exit(-1)
	}
}

func ReloadCfg( cfgName string)(error){
	cfgItem, exist := allCfg[cfgName];
	if !exist{
		err := fmt.Errorf("ReloadCfg:%v not exist!", cfgName)
		mylog.Error(err)
		return err
	}

	tmpCfg := reflect.New(reflect.TypeOf(cfgItem.cfgDataRef).Elem())
	if err := reloadCfgReal(cfgItem.fileName, tmpCfg.Interface()); err !=nil {
		return err
	}

	reflect.ValueOf(cfgItem.cfgDataRef).Elem().Set(tmpCfg.Elem())
	return nil
}

func reloadCfgReal(cfgFile string, outRef interface{}) error {
	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		mylog.Error("json file :" + cfgFile + " read failed:", err)
		fmt.Println("server start error:", cfgFile, " json file read failed:", err)
		return err
	}
	err = json.Unmarshal(data, outRef)
	if err != nil {
		mylog.Error("ExCfg json unmarshal failed:", err)
		fmt.Println("server start error:", cfgFile, " json unmarshal failed,", err)
		return err
	}

	//检查是否有成员仍然是默认值
	canNotSetFiledNames := make([]string, 0)
	outType := reflect.TypeOf(outRef).Elem()
	outValues := reflect.ValueOf(outRef).Elem()
	for i := 0; i < outType.NumField(); i++ {
		filedType := outType.Field(i)
		fieldValue := outValues.Field(i)
		if reflect.DeepEqual( fieldValue.Interface(), reflect.New(filedType.Type).Elem().Interface()){//
			canNotSetFiledNames = append(canNotSetFiledNames, filedType.Name)
		}
	}
	if len(canNotSetFiledNames)>0{
		outInfo := fmt.Errorf("filed %v is default value in config file:\"%v\"", canNotSetFiledNames, cfgFile)
		mylog.Warn(outInfo)
		fmt.Println("WARN:", outInfo)
	}

	return nil
}
