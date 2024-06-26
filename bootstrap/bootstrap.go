/*
@Author  :   xff
@Desc    :	 bootstrap
*/

package bootstrap

import (
	"encoding/json"
	"fmt"
	"gAudit/config"
	"gAudit/global"
	"gAudit/middleware"
	"gAudit/models"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// InitializeLog 初始化日志
func InitializeLog() {
	global.App.Log = middleware.InitLogger("app.log")
}

// InitializeDB 初始化数据库
func InitTables() {
	err := global.App.DB.AutoMigrate(&models.TestCase{})
	if err != nil {
		global.App.Log.Fatal(err.Error())
	}
}

// InitializeAuditConfig 初始化AuditConfig
func InitializeAuditConfig(configFile string) *config.AuditConfiguration {
	var AuditConfig = config.NewAuditConfiguration()

	// 读取JSON配置文件
	file, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)

	// 将配置文件值赋值给初始化默认值的AuditConfig
	err = decoder.Decode(AuditConfig)
	if err != nil {
		panic(err)
	}

	return AuditConfig
}

// InitializeDB 初始化数据库
func InitDB(user string, password string, host string, port int, database string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return db, err
	}
	err = db.Use(
		dbresolver.Register(dbresolver.Config{ /* xxx */ }).
			SetConnMaxIdleTime(600 * time.Second).
			SetConnMaxLifetime(600 * time.Second).
			SetMaxIdleConns(64).
			SetMaxOpenConns(64),
	)
	if err != nil {
		return db, err
	}
	return db, err
}
