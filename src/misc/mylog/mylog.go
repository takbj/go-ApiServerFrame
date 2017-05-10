package mylog

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"misc/timefix"

	"github.com/alecthomas/repr"
)

const (
	LOG_DEBUG = 3
	LOG_INFO  = 2
	LOG_WARN  = 1
	LOG_ERROR = 0
)

var PRINT = true
var mylog *log.Logger
var logchan chan string
var logFile *os.File
var LogLevel int = 100
var appName string

func MyLogger() *log.Logger {
	return mylog
}

func init() {
	appName = strings.Replace(os.Args[0], "\\", "/", -1)
	_, name := path.Split(appName)
	names := strings.Split(name, ".")
	appName = names[0]

	fileName := "./log/" + appName + time.Now().Format("20060102") + ".log"
	var err error
	logFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("open log file error !", fileName)
		os.Exit(-1)
		return
	}
	logchan = make(chan string, 1000)
	mylog = log.New(logFile, "", log.Ldate|log.Ltime)

	go writeloop()
}

func Debug(v ...interface{}) {
	if LogLevel > LOG_DEBUG {
		logchan <- "[DEBUG] " + fmt.Sprint(v)
	}
}

func Info(v ...interface{}) {
	if LogLevel > LOG_INFO {
		logchan <- "[INFO] " + fmt.Sprint(v)
	}
}

func Warn(v ...interface{}) {
	if LogLevel > LOG_WARN {
		logchan <- "[WARN] " + fmt.Sprint(v)
	}
}

func Error(v ...interface{}) {
	logchan <- "[ERROR] " + fmt.Sprint(v)
	logchan <- PrintStack()
}

func Println(v interface{}) {
	if !PRINT || v == nil {
		return
	}
	repr.Println(v, repr.Indent(" "))
}

func PrintStack() (ret string) {
	for i := 0; i < 15; i++ {
		funcName, file, line, ok := runtime.Caller(i)
		if ok {
			ret += fmt.Sprintf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
		}
	}
	return ret
}

func PrintPanicStack() {
	if x := recover(); x != nil {
		Error(fmt.Sprintf("%v", x))
		// PrintStack()
	}
}

func writeloop() {
	pm := time.NewTimer(time.Duration(timefix.NextMidnight(time.Now(), 1).Unix()-time.Now().Unix()) * time.Second)
	for {
		select {
		case str := <-logchan:
			mylog.Println(str)
		case <-pm.C:
			// 关闭原来的日志文件
			logFile.Close()

			time.Sleep(time.Second * 1)

			fileName := "./log/" + appName + time.Now().Format("20060102") + ".log"
			var err error
			logFile, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalln("open log file error !", fileName)
				os.Exit(-1)
				return
			}

			mylog.SetOutput(logFile)

			pm.Reset(time.Second * 24 * 60 * 60)
		}
	}
}
