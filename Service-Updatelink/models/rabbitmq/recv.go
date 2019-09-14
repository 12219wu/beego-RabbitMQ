package rabbitmq

import (
	"GridService/Service-Updatelink/logs"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/astaxie/beego"

	"github.com/streadway/amqp"
)

type ExchangeIndex struct {
	GridId string `json:"gridId"`
	Count  int32  `json:"count"`
}

var (
	rabbitMQAddr              string
	rabbitMQQueueName         string
	rabbitMQExchangeName      string
	rabbitMQExchangeNameToWeb string
	hchFileName               string
	debugOn                   bool
)

func init() {
	debugOn, _ = beego.AppConfig.Bool("debugOn")
	rabbitMQAddr = beego.AppConfig.String("rabbitMQAddr")
	rabbitMQQueueName = beego.AppConfig.String("rabbitMQQueueName")
	rabbitMQExchangeName = beego.AppConfig.String("rabbitMQExchangeName")
	rabbitMQExchangeNameToWeb = beego.AppConfig.String("rabbitMQExchangeNameToWeb")
	hchFileName = beego.AppConfig.String("hchFileName")
	logs.Info.Println("RabbitMQ init")
	logs.Info.Println("rabbitMQ登录地址:", rabbitMQAddr)
	logs.Info.Println("rabbitMQ交换器队列名称:", rabbitMQExchangeName)
	logs.Info.Println("rabbitMQ消息队列名称:", rabbitMQQueueName)
	go RabbitMQDataRoutine()
}

//处理rabbitMQ消息队列数据
func RabbitMQDataRoutine() {
	go RabbitMQRecvContinue()
}

//临时变量，保存从消息队列中接收到的值
type TempMsg struct {
	TimeStamp       int64
	ExchangeDataArr []ExchangeIndex
}

var TempMsgArr []TempMsg

//添加读写锁
var RW sync.RWMutex

//实时读取rabbitMQ队列消息，并将读取到的数据存到TempMsgArr中
func RabbitMQRecvContinue() (error, []TempMsg) {
	logs.Info.Println("进入RabbitMQRecvContinue线程函数...")
	conn, err := amqp.Dial(rabbitMQAddr)
	if err != nil {
		logs.Error.Printf("amqp.Dial错误: %s\n", err.Error())
		return err, TempMsgArr
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		logs.Error.Printf("conn.Channel错误: %s\n", err.Error())
		return err, TempMsgArr
	}
	defer ch.Close()

	args := amqp.Table{"x-message-ttl": int32(50000)}
	err = ch.ExchangeDeclare(
		rabbitMQExchangeName, // name of the exchange
		"fanout",             // type
		true,                 // durable
		false,                // delete when complete
		false,                // internal
		false,                // noWait
		nil,                  // arguments
	)
	if err != nil {
		logs.Error.Printf("ch.ExchangeDeclare错误: %s\n", err.Error())
		return err, TempMsgArr
	}
	//args := amqp.Table{"x-message-ttl": int32(10000)}
	q, err := ch.QueueDeclare(
		rabbitMQQueueName, //name
		false,             //durable
		false,             //delete when unused
		false,             //exclusive
		false,             //no-wait
		args,              //arguments
	)
	if err != nil {
		logs.Error.Printf("ch.QueueDeclare错误: %s\n", err.Error())
		return err, TempMsgArr
	}
	err = ch.QueueBind(
		q.Name,
		"",
		rabbitMQExchangeName,
		false,
		nil,
	)
	if err != nil {
		logs.Error.Printf("ch.QueueBind错误: %s\n", err.Error())
		return err, TempMsgArr
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		logs.Error.Printf("ch.Consume错误: %s\n", err.Error())
		return err, TempMsgArr
	}
	//时间戳标记
	timeStamp := time.Now().Unix()
	forever := make(chan bool)
	//defer close(forever)
	go func(ch chan bool) {
		for d := range msgs {
			var tempMsg TempMsg
			var recvArr []ExchangeIndex
			err = json.Unmarshal(d.Body, &recvArr)
			if err != nil {
				log.Fatalf("json.Unmarshal错误:%s\n!", err.Error())
			}
			tempMsg.ExchangeDataArr = recvArr
			tempMsg.TimeStamp = timeStamp
			timeStamp++
			RW.Lock()
			TempMsgArr = append(TempMsgArr, tempMsg)
			RW.Unlock()
		}
	}(forever)
	logs.Info.Println("RabbitMQRecvContinue开启,等待接收消息...")
	<-forever
	return nil, TempMsgArr
}

//获取存储接收rabbitMQ的消息的变量TempMsgArr
func GetRabbitMQTemp() (error, []ExchangeIndex, int64, int64) {
	RW.Lock()
	logs.Info.Println("当前缓存数据大小:", len(TempMsgArr))
	timeStamp := int64(0)
	for _, val := range TempMsgArr {
		if val.TimeStamp >= timeStamp {
			timeStamp = val.TimeStamp
		}
	}
	var exchangeIndexArr []ExchangeIndex
	var arrIndex, tempMsgArrLength int64
	tempMsgArrLength = int64(len(TempMsgArr))
	for _, val := range TempMsgArr {
		if val.TimeStamp == timeStamp {
			exchangeIndexArr = val.ExchangeDataArr
			arrIndex = val.TimeStamp
		}
	}
	//TempMsgArr = nil
	RW.Unlock()
	return nil, exchangeIndexArr, arrIndex, tempMsgArrLength
}

//清空临时变量	TempMsgArr
func ClearTempMsgArr() {
	if len(TempMsgArr) != 0 {
		TempMsgArr = nil
	}
}

//缓存数据之前，先清理之前rabbitMQ队列
func ClearRabbitMQQueueData() (error, int64) {
	conn, err := amqp.Dial(rabbitMQAddr)
	if err != nil {
		logs.Error.Printf("amqp.Dial错误: %s\n", err.Error())
		return err, 0
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		logs.Error.Printf("conn.Channel错误: %s\n", err.Error())
		return err, 0
	}
	defer ch.Close()
	err = ch.ExchangeDeclare(
		rabbitMQExchangeName, // name of the exchange
		"fanout",             // type
		true,                 // durable
		false,                // delete when complete
		false,                // internal
		false,                // noWait
		nil,                  // arguments
	)
	if err != nil {
		logs.Error.Printf("ch.ExchangeDeclare错误: %s\n", err.Error())
		return err, 0
	}
	q, err := ch.QueueDeclare(
		rabbitMQQueueName, // name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		logs.Error.Printf("ch.QueueDeclare错误: %s\n", err.Error())
		return err, 0
	}
	err = ch.QueueBind(
		q.Name,
		"",
		rabbitMQExchangeName,
		false,
		nil,
	)
	if err != nil {
		logs.Error.Printf("ch.QueueBind错误: %s\n", err.Error())
		return err, 0
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		logs.Error.Printf("ch.Consume错误: %s\n", err.Error())
		return err, 0
	}
	//时间戳标记
	//timeStamp := time.Now().Unix()
	stop := make(chan bool)
	//记录清理的消息条数
	var numOfMsg int64
	go func() {
		for d := range msgs {
			select {
			case <-stop:
				return
			default:
				numOfMsg++
				_ = d.Body
			}
		}
	}()
	//清理数据时间
	clearRabbitMQtime, _ := beego.AppConfig.Int("clearRabbitMQtime")
	time.Sleep(time.Duration(clearRabbitMQtime) * time.Second)
	logs.Info.Println("清理数据耗时: ", clearRabbitMQtime, "秒")
	stop <- true
	logs.Info.Println("清理数据条数numOfMsg=", numOfMsg)
	return nil, numOfMsg
}
