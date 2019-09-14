package logs

import (
	"github.com/astaxie/beego"
	"io"
	"log"
	"os"
	"time"
)

var (
	logOn   bool
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	logOn,_ = beego.AppConfig.Bool("logOn")
	fileDir := "static/log/"
	err := os.MkdirAll(fileDir,0777)
	if err!=nil{
		log.Fatalln("创建日志文件目录失败：", err)
	}
	fileNameByTime := fileDir + time.Now().Format("20060102150405") + ".log"
	errFile, err := os.OpenFile(fileNameByTime, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	var writer io.Writer
	if logOn {
		writer = io.MultiWriter(os.Stdout,errFile)
	}else{
		writer = io.Writer(os.Stdout)
	}
	Info = log.New(writer, "Info:", log.Ldate|log.Ltime|log.Llongfile)
	Warning = log.New(writer, "Warning:", log.Ldate|log.Ltime|log.Llongfile)
	Error = log.New(writer, "Error:", log.Ldate|log.Ltime|log.Llongfile)
}
func MyExample() {
	Info.Println("飞雪无情的博客:", "http://www.flysnow.org")
	Warning.Printf("飞雪无情的微信公众号：%s\n", "flysnow_org")
	Error.Println("欢迎关注留言")
	//打印结果
	//Info:2019/07/19 11:09:07 E:/GoWS/src/go_log/main.go:30: 飞雪无情的博客: http://www.flysnow.org
	//Warning:2019/07/19 11:09:07 E:/GoWS/src/go_log/main.go:31: 飞雪无情的微信公众号：flysnow_org
	//Error:2019/07/19 11:09:07 E:/GoWS/src/go_log/main.go:32: 欢迎关注留言
}