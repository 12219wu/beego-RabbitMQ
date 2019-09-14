package models

import (
	"GridService/Service-Spacdt/models/spacbasic"
	"GridService/Service-Updatelink/logs"
	"GridService/Service-Updatelink/models/rabbitmq"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"xh.common/xh_util"
)

var (
	memCache         xh_util.StMemcached
	threshold        int
	lockKey          string
	polyhedronTbName string
	routeLinkTbName  string
	sqlName          string
	sqlHost          string
	sqlUser          string
	sqlPassword      string
	sqlPort          int
	debugOn          bool
	logOn            bool
	//db.User=root
	//db.Pwd=123456
	//db.Host=10.100.1.140
	//db.Port=3306
	//db.Name=grid_map
	//const char* Name, const char* Host, const char* User, const char* Password, unsigned int Port)
)

func init() {
	logOn, _ = beego.AppConfig.Bool("logOn")
	debugOn, _ = beego.AppConfig.Bool("debugOn")
	sqlName = beego.AppConfig.String("db.Name")
	sqlHost = beego.AppConfig.String("db.Host")
	sqlUser = beego.AppConfig.String("db.User")
	sqlPassword = beego.AppConfig.String("db.Pwd")
	sqlPort, _ = beego.AppConfig.Int("db.Port")
	polyhedronTbName = beego.AppConfig.String("polyhedronTbName")
	routeLinkTbName = beego.AppConfig.String("routeLinkTbName")
	threshold, _ = beego.AppConfig.Int("Threshold")
	logs.Info.Println("当前人流密度阈值: ",threshold)
	lockIp := beego.AppConfig.String("memcache_host")
	lockPort := beego.AppConfig.String("memcache_port")
	lockKey = beego.AppConfig.String("memcache_key")
	memCache.Init(lockIp, lockPort)
	go msgRoutine()
}

func msgRoutine() {
	for {
		if len(rabbitmq.TempMsgArr) != 0 {
			updateLinkStart := time.Now().UnixNano() / 1e6
			logs.Info.Println("执行UpdateLinkTable更新link数据表!")
			//将从rabbitMQ接收到的人流密度数据转换为link写入数据库
			err := UpdateLinkTable()
			if err != nil {
				logs.Error.Println("msgRoutine:UpdateLinkTable失败:", err.Error())
				continue
			}
			updateLinkEnd := time.Now().UnixNano() / 1e6
			logs.Info.Println("UpdateLinkTable更新link数据表完毕!耗时:", updateLinkEnd-updateLinkStart, "毫秒")
			//将从rabbitMQ接收到的人流密度数据转换为经纬度数据发送给前端,与更新link用的是同一份数据
			sendToWebStart := time.Now().UnixNano() / 1e6
			err = rabbitmq.RabbitMQSendLonLat()
			if err != nil {
				logs.Error.Println("msgRoutine:RabbitMQSendLonLat失败:", err.Error())
				continue
			}
			sendToWebEnd := time.Now().UnixNano() / 1e6
			logs.Info.Println("sendToWeb耗时:", sendToWebEnd-sendToWebStart, "毫秒")
			rabbitmq.ClearTempMsgArr()
		} else {
			if debugOn {
				logs.Info.Println("等待接收数据...")
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func UpdateLinkTable() error {
	var int4mIndexs []int64
	var int1mIndexs []int64
	err, recvArr, _, recvArrLen := rabbitmq.GetRabbitMQTemp()
	if err != nil {
		logs.Error.Printf("UpdateLinkTable中RabbitMQRecv错误:%s!\n", err.Error())
		return err
	}
	if recvArrLen == 0 {
		logs.Error.Println("UpdateLinkTable:GetRabbitMQTemp返回数据为空!")
		return errors.New("UpdateLinkTable:GetRabbitMQTemp返回数据为空!")
	}
	if debugOn {
		sec := xh_util.GetCurrentTime()
		bytes, _ := json.Marshal(&recvArr)
		fileName := fmt.Sprintf("static/CrowdDensity/%d.txt", sec)
		ioutil.WriteFile(fileName, bytes, os.ModePerm)
	}
	for _, val := range recvArr {
		//人流密度阈值
		if val.Count >= int32(threshold) {
			intIndex, err := strconv.ParseInt(val.GridId, 10, 64)
			if err != nil {
				logs.Error.Printf("UpdateLinkTable中ParseInt错误:%s!\n", err.Error())
				return err
			}
			int4mIndexs = append(int4mIndexs, intIndex)
		}
	}

	for _, val := range int4mIndexs {
		myIndexArr := spacbasic.NativeGetProgenies(val, 2)
		//myIndexArr中有64个网格，只取z=41的网格，总共所需的为16个
		index41Arr := myIndexArr[16:32]
		int1mIndexs = append(int1mIndexs, index41Arr...)
	}
	if logOn {
		logs.Info.Println("UpdateLinkTable筛选后的int4mIndexs大小为:", len(int4mIndexs))
		logs.Info.Println("UpdateLinkTable筛选后的int4mIndexs:", int4mIndexs)
		logs.Info.Println("4m的格子转换成1米的格子大小为: ", len(int1mIndexs))
		logs.Info.Println("4m的格子转换成1米的格子: ", int1mIndexs)
	}
	//删除实时数据表中上次残留数据
	o := orm.NewOrm()
	//删除网格数据
	deletePolyhedronStart := time.Now().UnixNano() / 1e6
	num, err := o.QueryTable(polyhedronTbName).Filter("map_id", 2000).Delete()
	if err != nil {
		logs.Error.Printf("UpdateLinkTable中Delete错误:%s!\n", err.Error())
		return err
	}
	deletePolyhedronEnd := time.Now().UnixNano() / 1e6
	logs.Info.Println("删除网格数据耗时毫秒: ", deletePolyhedronEnd-deletePolyhedronStart)
	logs.Info.Printf("清除%s上次残留数据条数:%d\n", polyhedronTbName, num)
	//数据库表更新
	//添加网格数据,插入到基础障碍物网格数据表中从id=500000开始
	var addIndexs []*GridIndex
	myNum := 500000
	for _, val := range int1mIndexs {
		var addIndex GridIndex
		addIndex.Id = int32(myNum)
		myNum++
		addIndex.MapId = 2000
		addIndex.SpaceIndex = val
		addIndexs = append(addIndexs, &addIndex)
	}
	addPolyhedronStart := time.Now().UnixNano() / 1e6
	if len(addIndexs) != 0 {
		insertNums, err := o.InsertMulti(5000, addIndexs)
		if err != nil {
			logs.Error.Println("UpdateLinkTable插入数据失败！", err.Error())
			return err
		}
		logs.Info.Printf("%s中插入数据总数%d\n", polyhedronTbName, insertNums)
	} else {
		logs.Info.Printf("%s中插入数据总数为0\n", polyhedronTbName)
	}
	addPolyhedronEnd := time.Now().UnixNano() / 1e6
	logs.Info.Println("添加网格数据耗时毫秒: ", addPolyhedronEnd-addPolyhedronStart)
	//数据库加锁
	err = memCache.Lock(lockKey, 50)
	if err != nil {
		logs.Error.Println("UpdateLinkTable->memCache.Lock返回错误:", err.Error())
	}

	//删除link数据
	deleteLinkStart := time.Now().UnixNano() / 1e6
	num, err = o.QueryTable(routeLinkTbName).Filter("area_id", 2000).Delete()
	if err != nil {
		logs.Error.Printf("UpdateLinkTable中Delete错误:%s!\n", err.Error())
		memCache.UnLock(lockKey)
		return err
	}
	deleteLinkEnd := time.Now().UnixNano() / 1e6
	logs.Info.Println("删除link数据耗时: ", deleteLinkEnd-deleteLinkStart)
	logs.Info.Printf("清除%s上次残留数据条数:%d\n", routeLinkTbName, num)
	linkStart := time.Now().UnixNano() / 1e6
	//更新link表
	err = spacbasic.NativeTransroute(polyhedronTbName, routeLinkTbName, 0, 2000, sqlName, sqlHost, sqlUser, sqlPassword, uint(sqlPort))
	linkEnd := time.Now().UnixNano() / 1e6
	logs.Info.Println("调用DLL转换link耗时毫秒: ", linkEnd-linkStart)
	memCache.UnLock(lockKey)
	if err != nil {
		logs.Error.Println("UpdateLinkTable调用spacbasic.NativeTransroute失败！", err.Error())
	}
	return err
}
