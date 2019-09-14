package main

import (
	_ "GridService/Service-RabbitMQSend/routers"
	_ "GridService/Service-RabbitMQSend/models"
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}

