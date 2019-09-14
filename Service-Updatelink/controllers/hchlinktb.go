package controllers

import (
	"GridService/Service-Spacdt/models/spacbasic"
	
	"GridService/Service-Updatelink/models"
	"GridService/Service-Updatelink/models/rabbitmq"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"time"
	"xh.common/xh_util"
)

type UpdateLinkServiceController struct {
	beego.Controller
}

type TagCode struct {
	Code int
	Msg  string
	Data interface{}
}

func init() {
	spacbasic.NativeInit()
}

func (c *UpdateLinkServiceController) RbmqSend() {
	var tag TagCode
	err := rabbitmq.RabbitMQSend()
	if err!=nil{
		tag.Code = 1
		tag.Msg = "RbmqSend RabbitMQSend失败!"
		tag.Data = err.Error()
		c.Data["json"] = tag
		c.ServeJSON()
		return
	}
	tag.Code = 0
	tag.Msg = "RbmqSend成功!"
	c.Data["json"] = tag
	c.ServeJSON()
	return
}

func (c *UpdateLinkServiceController) RbmqRecv() {
	var tag TagCode
	err,strIndexs,arrIndex,tempMsgArrLength := rabbitmq.GetRabbitMQTemp()
	if err != nil {
		tag.Code = 1
		tag.Msg = "RbmqRecv RabbitMQRecv失败!"
		tag.Data = err.Error()
		c.Data["json"] = tag
		c.ServeJSON()
		return
	}
	tag.Code = 0
	tag.Msg = fmt.Sprintf("arrIndex:%d len(strIndexs):%d tempMsgArrLength:%d",arrIndex,len(strIndexs),tempMsgArrLength)
	tag.Data = strIndexs
	c.Data["json"] = tag
	c.ServeJSON()
	return
}

func (c *UpdateLinkServiceController) UpdateLinkTb() {
	var tag TagCode
	start := time.Now().Unix() //获取时间戳
	err := models.UpdateLinkTable()
	if err != nil {
		tag.Code = 1
		tag.Msg = "UpdateLinkTb UpdateLinkTable失败!"
		c.Data["json"] = tag
		c.ServeJSON()
		return
	}
	end := time.Now().Unix() //获取时间戳
	tag.Code = 0
	tag.Msg = "UpdateLinkTb成功!"
	tag.Data = end - start
	c.Data["json"] = tag
	c.ServeJSON()
	return
}
//获取后代网格
type FatherIndex struct {
	GridIndex  int64  `json:"gridindex"`
	Generation uint32 `json:"generation"`
}
func (c *UpdateLinkServiceController) GetProgenies() {
	fatherIndex := FatherIndex{}
	var tag TagCode
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &fatherIndex)
	if err != nil {
		tag.Code = 1
		tag.Msg = "GetGetProgenies Unmarshal失败!"
		c.Data["json"] = tag
		c.ServeJSON()
		return
	}
	fmt.Println(fatherIndex)
	myIndexArr := spacbasic.NativeGetProgenies(fatherIndex.GridIndex, fatherIndex.Generation)
	//index41Arr := myIndexArr[16:32]
	tag.Code = 0
	tag.Msg = "查询后代网格索引成功!"
	tag.Data = myIndexArr
	c.Data["json"] = tag
	c.ServeJSON()
	return
}

type NeighborIndex struct {
	Index int64 `json:"index"`
	Dimension int `json:"dimension"`
}

//获取邻居网格
func (c *UpdateLinkServiceController) GetNeighborsGrid() {
	neighborIndex := NeighborIndex{}
	var tag TagCode
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &neighborIndex)
	if err != nil {
		tag.Code = 1
		tag.Msg = "GetNeighborsGrid Unmarshal失败!"
		c.Data["json"] = tag
		c.ServeJSON()
		return
	}
	myArr := spacbasic.NativeGetNeighbors(neighborIndex.Index,neighborIndex.Dimension)
	//myArr := GetNeibors8Indexs(neighborIndex.Index,neighborIndex.Dimension)
	tag.Code = 0
	tag.Msg = "GetNeighborsGrid成功!"
	tag.Data = myArr
	c.Data["json"] = tag
	c.ServeJSON()
	return
}

//验证cgo
type StLevelPoint3D struct {
	Level int32             `json:"level"`
	Point xh_util.StPoint3D `json:"point"`
}
func (c *UpdateLinkServiceController) MakeSpatialIndexBoxByPos() {
	pos := StLevelPoint3D{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &pos)
	if err != nil {
		xh_util.XhError("MakeSpatialIndexBoxByPos param err")
		ret := xh_util.ErrorInstance(xh_util.PARAM_ERROR)
		c.Data["json"] = &ret
		c.ServeJSON()
		return
	}

	box := xh_util.StBox3D{}
	spacbasic.NativeMakeSpatialIndexBoxByPos(pos.Point, pos.Level, &box)
	xh_util.XhInfo("MakeSpatialIndexBoxByPos ", box)

	c.Data["json"] = &box
	c.ServeJSON()
}