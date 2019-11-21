package logger

import (
	filetool "github.com/olegpolukhin/go_ps_scraping/file"
	"os"
	"strconv"
	"time"
)

var logFile *os.File
var logInited bool = false

// LogFile name file
const LogFile = "logs.log"

func Init() {
	logFile = filetool.CreateFile(LogFile)
	filetool.AppendToFile(logFile, getDateTime()+"Log started\n")
	filetool.CloseFile(logFile)
	logInited = true
}

func Write(logMessage string) {
	if logInited {
		logFile = filetool.OpenFile(LogFile)
		filetool.AppendToFile(logFile, getDateTime()+logMessage+"\n")
		filetool.CloseFile(logFile)
	} else {
		Init()
		logFile = filetool.OpenFile(LogFile)
		filetool.AppendToFile(logFile, getDateTime()+logMessage+"\n")
		filetool.CloseFile(logFile)
	}
}

func getDateTime() string {
	var hour = time.Now().Hour()
	var minute = time.Now().Minute()
	var second = time.Now().Second()
	var yy, mm, dd = time.Now().Date()
	return strconv.Itoa(dd) + "." + mm.String() + "." + strconv.Itoa(yy) + "_" + strconv.Itoa(hour) + ":" + strconv.Itoa(minute) + ":" + strconv.Itoa(second) + "_|_"
}
