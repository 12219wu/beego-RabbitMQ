package models

import (
	"GridService/Service-Spacdt/models/spacbasic"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/pkg/errors"
	"xh.common/xh_util"

	_ "github.com/astaxie/beego"
	"github.com/streadway/amqp"
)

func init() {
	spacbasic.NativeInit()
}

var (
	rabbitMQAddr         string
	rabbitMQExchangeName string
	hchFileName          string
)

func init() {
	rabbitMQAddr = beego.AppConfig.String("rabbitMQAddr")
	fmt.Println(rabbitMQAddr)
	rabbitMQExchangeName = beego.AppConfig.String("rabbitMQExchangeName")
	fmt.Println(rabbitMQExchangeName)
	hchFileName = beego.AppConfig.String("hchFileName")
	fmt.Println(hchFileName)
	go sendRoutine()
}

func sendRoutine() {
	for {
		go RabbitMQSend()
		time.Sleep(20 * time.Second)
	}
}

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
	//	err, bytes := Trans2LonLatArr()
	//	if err != nil {
	//		return err
	//	}
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
	fmt.Printf("Send %s\n", string(bytes))
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

type LonLat struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type GridCrow struct {
	GridId string `json:"gridId"`
	Count  int64  `json:"count"`
}

func Trans2LonLatArr() (error, []byte) {
	var gridCrowArr []GridCrow
	var lonLatArr []LonLat
	err, bytes := ReadFromTxt()
	if err != nil {
		fmt.Println(err.Error())
		return err, nil
	}
	err = json.Unmarshal(bytes, &gridCrowArr)
	if err != nil {
		fmt.Println(err.Error())
		return err, nil
	}
	for _, val := range gridCrowArr {
		var box xh_util.StBox3D
		var lonLat LonLat
		gridID, err := strconv.ParseUint(val.GridId, 10, 64)
		if err != nil {
			fmt.Println(err.Error())
			return err, nil
		}
		b := spacbasic.NativeMakeSpatialIndexBox(gridID, &box)
		if !b {
			fmt.Println(b)
			return errors.New("Trans2LonLatArr:spacbasic.NativeMakeSpatialIndexBox失败!"), nil
		}
		lonLat.Longitude = (box.SouthWest.X + box.NorthEast.X) / 2
		lonLat.Latitude = (box.SouthWest.Y + box.NorthEast.Y) / 2
		for i := 0; i < int(val.Count); i++ {
			lonLatArr = append(lonLatArr, lonLat)
		}
	}
	lonLatArrBytes, err := json.Marshal(lonLatArr)
	if err != nil {
		fmt.Println(err.Error())
		return err, nil
	}
	return nil, lonLatArrBytes
}
