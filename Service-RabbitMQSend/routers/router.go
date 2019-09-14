package routers

import (
	"GridService/Service-RabbitMQSend/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
