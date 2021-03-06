package models

import (
	"fmt"
	"github.com/coder2z/g-saber/xcfg"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

var (
	MysqlHandler *gorm.DB
)

func mysqlBuild() *gorm.DB {
	var err error
	DB, err := gorm.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
			xcfg.GetString("db.name"),
			xcfg.GetString("db.pw"),
			xcfg.GetString("db.addr"),
			xcfg.GetString("db.dbname")))

	if err != nil {
		log.Panicf("models.Setup err: %v", err)
		return nil
	}

	//	设置表前缀
	//gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
	//	return setting.DatabaseSetting.TablePrefix + defaultTableName
	//}

	DB.SingularTable(true)
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	return DB
}

func Init() {
	MysqlHandler = mysqlBuild()
}
