/*
@Time    :   2022/07/06 10:07:46
@Author  :   zongfei.fu
@Desc    :   暂时没用
*/

package models

import (
	"gorm.io/gorm"
)

// 本地测试用例表
type TestCase struct {
	gorm.Model
	Env           string `gorm:"type:varchar(32) not null default '';comment '环境'" json:"env"`
	ClusterName   string `gorm:"type:varchar(128) not null default '';comment:'集群名'" json:"cluster_name"`
	Datacenter    string `gorm:"type:varchar(128) not null default '';comment:'数据中心'" json:"datacenter"`
	Region        string `gorm:"type:varchar(128) not null default '';comment:'区域'" json:"region"`
	Hostname      string `gorm:"type:varchar(128) not null default '';comment:'主机名'" json:"hostname"`
	Port          string `gorm:"type:int(11) not null default 3306;comment:'端口'" json:"port"`
	PromotionRule string `gorm:"type:varchar(128) not null default 'prefer';comment:'晋升规则,可选值:prefer/neutral/prefer_not/must_not'" json:"promotion_rule"`
}

func (TestCase) TableName() string {
	return "test_case"
}
