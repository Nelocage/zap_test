package store

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Db *gorm.DB

func InitGorm() {
	dsn := "root:123456@tcp(172.22.114.78:3306)/topcloud?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{
		SingularTable: true, // 使用单数表名，启用该选项，此时，`Article` 的表名应该是 `it_article`
	}})
	Db = db
	if err != nil {
		fmt.Println("数据库连接失败", err)
	}
}

type DB interface {
	Get(key string) (int, error)
}

func GetFromDB(db DB, key string) int {
	if value, err := db.Get(key); err == nil {
		return value
	}

	return -1
}
