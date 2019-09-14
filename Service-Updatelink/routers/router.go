package routers

import (
	"GridService/Service-Updatelink/controllers"

	"github.com/astaxie/beego"

	"github.com/astaxie/beego/plugins/cors"
)

func init() {
	// 支持跨域访问，暂时先这么处理
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))
	beego.Router("/grid/updatelink/getprogenies", &controllers.UpdateLinkServiceController{}, "post:GetProgenies")
	beego.Router("/grid/updatelink/getneighborsgrid", &controllers.UpdateLinkServiceController{}, "post:GetNeighborsGrid")
	//rabbitMQ发送数据
	beego.Router("/grid/updatelink/rbmqsend", &controllers.UpdateLinkServiceController{}, "get:RbmqSend")
	beego.Router("/grid/updatelink/rbmqrecv", &controllers.UpdateLinkServiceController{}, "get:RbmqRecv")
	//更新link表
	beego.Router("/grid/updatelink/updatelinktb", &controllers.UpdateLinkServiceController{}, "get:UpdateLinkTb")
	//验证cgo
	beego.Router("/grid/updatelink/makespatialindexboxbypos", &controllers.UpdateLinkServiceController{}, "post:MakeSpatialIndexBoxByPos")
	
}
