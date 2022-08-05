/*
@Time    :   2022/06/23 16:37:04
@Author  :   zongfei.fu
@Desc    :   操作本地数据库，暂时没用
*/

package models

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

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
