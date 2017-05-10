package utils

import (
	"fmt"
	"io"
	"time"

	"log"
	"os"

	"strings"

	// "misc/timefix"
	"path"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

var logFile *os.File
var appName string

//开启mysql数据库
func OpenServerDB(addr string, maxOpenConns, maxIdleConns int, debug, force bool, register func()) {
	orm.Debug = debug
	if orm.Debug {
		orm.DebugLog = orm.NewLog(mysqlDebugOut())
		go writeloop()
	}
	params := make([]int, 0, 2)
	if maxOpenConns > 0 {
		params = append(params, maxIdleConns)
		params = append(params, maxOpenConns)
	} else if maxIdleConns > 0 {
		params = append(params, maxIdleConns)
	}
	if err := orm.RegisterDataBase("default", "mysql", addr, params...); err != nil {
		panic(err)
	}
	register()

	err := orm.RunSyncdb("default", force, debug)
	if err != nil {
		panic(err)
	}
}

func mysqlDebugOut() io.Writer {
	appName = strings.Replace(os.Args[0], "\\", "/", -1)
	_, name := path.Split(appName)
	names := strings.Split(name, ".")
	appName = names[0]
	appName = fmt.Sprintf("%v_sqldebug", appName)
	fileName := "log/" + appName + time.Now().Format("20060102") + ".log"
	var err error
	logFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("mysqlDebugOut,  open log file error !", fileName)
		os.Exit(-1)
	}
	return logFile
}

func writeloop() {
	pm := time.NewTimer(time.Duration(NextMidnight(time.Now(), 1).Unix()-time.Now().Unix()) * time.Second)
	for {
		select {
		case <-pm.C:
			// 关闭原来的文件
			fileName := "../log/" + appName + time.Now().Format("20060102") + ".stat"
			if statFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err == nil {
				orm.DebugLog.SetOutput(statFile)
				logFile.Close()
				logFile = statFile
			}
			pm.Reset(time.Second * 24 * 60 * 60)
		}
	}
}
