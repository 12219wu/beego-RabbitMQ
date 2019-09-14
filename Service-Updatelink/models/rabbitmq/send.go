package rabbitmq

import (
	"GridService/Service-Spacdt/models/spacbasic"
	"GridService/Service-Updatelink/logs"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/astaxie/beego"
	"github.com/streadway/amqp"
	"io/ioutil"
	"strconv"
	"xh.common/xh_util"
)

type LonLat struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

//单条发送网格ID与人数
func RabbitMQSend() error {
	conn, err := amqp.Dial(rabbitMQAddr)
	if err != nil {
		fmt.Printf("amqp.Dial错误: %s\n", err.Error())
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("conn.Channel错误: %s\n", err.Error())
		return err
	}
	defer ch.Close()
	//声明交换器
	err = ch.ExchangeDeclare(
		rabbitMQExchangeName, // name
		"fanout",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		fmt.Printf("ch.ExchangeDeclare错误: %s", err.Error())
		return err
	}
	//待发送数据
	err, bytes := ReadFromTxt()
	if err != nil {
		return err
	}
	err = ch.Publish(
		rabbitMQExchangeName, // exchange
		"",                   // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bytes,
			//Expiration: "50000",        // 设置消息过期时间
		})
	if debugOn {
		fmt.Printf("Send %s\n", string(bytes))
	}
	if err != nil {
		fmt.Printf("ch.Publish错误: %s\n", err.Error())
		return err
	}
	return nil
}

func ReadFromTxt() (error, []byte) {
	filePath := "static/" + hchFileName
	//fmt.Println(filePath)
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err, nil
	}
	return err, bytes
}

//单条发送由网格ID与人数转换的经纬度数据
func RabbitMQSendLonLat() error {
	conn, err := amqp.Dial(rabbitMQAddr)
	if err != nil {
		logs.Error.Printf("amqp.Dial错误: %s\n", err.Error())
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logs.Error.Printf("conn.Channel错误: %s\n", err.Error())
		return err
	}
	defer ch.Close()
	//声明交换器
	err = ch.ExchangeDeclare(
		rabbitMQExchangeNameToWeb, // name
		"fanout",                  // type
		true,                      // durable
		false,                     // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		logs.Error.Printf("ch.ExchangeDeclare错误: %s", err.Error())
		return err
	}
	//待发送数据
	err, bytes := Trans2LonLatArr()
	if err != nil {
		return err
	}
	err = ch.Publish(
		rabbitMQExchangeNameToWeb, // exchange
		"",    // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bytes,
			//Expiration: "50000",        // 设置消息过期时间
		})
	if debugOn {
		logs.Info.Printf("Send %s\n", string(bytes))
	}
	if err != nil {
		logs.Error.Printf("ch.Publish错误: %s\n", err.Error())
		return err
	}
	return nil
}

func Trans2LonLatArr() (error, []byte) {
	var gridCrowArr []ExchangeIndex
	var lonLatArr []LonLat
	err, gridCrowArr, _, gridCrowArrLen := GetRabbitMQTemp()
	if err != nil {
		logs.Error.Println(err.Error())
		return err, nil
	}
	if gridCrowArrLen == 0 {
		logs.Error.Println("Trans2LonLatArr获取[]ExchangeIndex长度为零!")
		return errors.New("Trans2LonLatArr获取[]ExchangeIndex长度为零!"), nil
	}
	for _, val := range gridCrowArr {
		var box xh_util.StBox3D
		var lonLat LonLat
		gridID, err := strconv.ParseUint(val.GridId, 10, 64)
		if err != nil {
			logs.Error.Println(err.Error())
			return err, nil
		}
		b := spacbasic.NativeMakeSpatialIndexBox(gridID, &box)
		if !b {
			logs.Error.Println(b)
			return errors.New("Trans2LonLatArr:spacbasic.NativeMakeSpatialIndexBox失败!"), nil
		}
		lonLat.Longitude = (box.SouthWest.X + box.NorthEast.X) / 2
		lonLat.Latitude = (box.SouthWest.Y + box.NorthEast.Y) / 2
		for i := 0; i < int(val.Count); i++ {
			lonLatArr = append(lonLatArr, lonLat)
		}
	}
	logs.Info.Println("send to web length []ExchangeIndex", len(gridCrowArr))
	logs.Info.Println("send to web length []LonLat:", len(lonLatArr))
	lonLatArrBytes, err := json.Marshal(lonLatArr)
	if err != nil {
		logs.Error.Println(err.Error())
		return err, nil
	}
	return nil, lonLatArrBytes
}
