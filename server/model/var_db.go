package model

import (
	"fmt"
	"time"

	"github.com/lazyfury/go-web-start/server/config"
	"github.com/lazyfury/go-web-template/model"
	"github.com/lazyfury/go-web-template/tools/types"
	// _ 数据库驱动
	// _ "gorm.io/gorm/dialects/sqlite"
)

// 需要自动迁移的表
var autoMigrate = []interface{}{
	//user
	&User{},
	&WechatOauth{},
	&WechatMiniUser{},

	//article
	&Articles{},
	&ArticlesCate{},
	&ArticlesRec{},
	&ArticlesTag{},
	//ad
	&AdEvent{},
	&AdGroup{},
	&Ad{},
	//feedback
	&Feedback{},
	// message
	&Message{},
	&MessageTemplate{},

	// 订单
	// &Order{},
}

// DB DB
var DB *model.GormDB = &model.GormDB{}

// MysqlConn InitDB
func MysqlConn(dsn string) (err error) {
	fmt.Printf("数据库链接>>>准备>> %s \n", time.Now().Format(types.DefaultTimeLayout))
	err = DB.ConnectMysql(dsn)
	// db, err := gorm.Open("sqlite3", "config/database.db")
	if err != nil {
		return
	}

	// gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
	// 	return TableName(defaultTableName)
	// }

	DB.AutoMigrate(autoMigrate...)

	fmt.Printf("数据库链接>>>成功>> %s \n", time.Now().Format(types.DefaultTimeLayout))
	return nil
}

//TableName 拼接默认表名
func TableName(str string) (result string) {
	result = config.Global.TablePrefix + "_" + str
	return result
}
