package models

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

//动态创建link用到
type RouteLink struct {
	Id            int64  `orm:"column(id)"`
	AreaId        int64  `orm:"column(area_id)"`
	BlockId       int64  `orm:"column(block_id)"`
	NeighborCount int    `orm:"column(neighbor_count)"`
	Neighbors     string `orm:"column(neighbors)"`
}

func (tb *RouteLink) TableName() string {
	return beego.AppConfig.String("routeLinkTbName")
}

type GridIndex struct {
	Id         int32 `orm:"column(id)"`
	MapId      int64 `orm:"column(map_id)"`
	SpaceIndex int64 `orm:"column(space_index)"`
}

func (tb *GridIndex) TableName() string {
	return beego.AppConfig.String("polyhedronTbName")
}

func init() {
	fmt.Println("MySQL init")

	dbUser := beego.AppConfig.String("db.User")
	dbPwd := beego.AppConfig.String("db.Pwd")
	dbHost := beego.AppConfig.String("db.Host")
	dbPort := beego.AppConfig.String("db.Port")
	dbName := beego.AppConfig.String("db.Name")
	conn := dbUser + ":" + dbPwd + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8"
	fmt.Println(conn)
	// set default database
	orm.RegisterDataBase("default", "mysql", conn, 30)
	// wyd
	orm.RegisterModel(new(RouteLink))
	orm.RegisterModel(new(GridIndex))
	// create table
	orm.RunSyncdb("default", false, true)
}
