package main

import (
	_ "GridService/Service-Updatelink/models"
	_ "GridService/Service-Updatelink/routers"
	
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}
