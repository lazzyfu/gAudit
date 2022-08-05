/*
@Time    :   2022/07/06 10:07:46
@Author  :   zongfei.fu
@Desc    :   暂时没用
*/

package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// 表config，这里没用
type Config struct {
	gorm.Model
	Name    string `gorm:"type:varchar(128) not null;uniqueIndex:uniq_name" json:"name"`
	Value   datatypes.JSON
	Comment string `gorm:"type:varchar(256);default:null;" json:"comment"`
}

func (Config) TableName() string {
	return "config"
}
