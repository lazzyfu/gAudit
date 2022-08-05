/*
@Time    :   2022/07/06 10:12:48
@Author  :   zongfei.fu
@Desc    :   None
*/

package process

import (
	"fmt"
	"sqlSyntaxAudit/global"

	"github.com/pingcap/parser/mysql"
)

type PrimaryKey struct {
	Table            string // 表名
	Column           string // 列
	Tp               byte   // 类型
	Flag             uint   // flag
	HasNotNull       bool   // 是否not null
	HasAutoIncrement bool   // 是否自增
}

func (p *PrimaryKey) CheckBigint() error {
	if p.Tp != mysql.TypeLonglong && global.App.AuditConfig.CHECK_PRIMARYKEY_USE_BIGINT {
		// 必须使用bigint类型
		return fmt.Errorf("表`%s`的主键%s必须使用bigint类型", p.Table, p.Column)
	}
	return nil
}

func (p *PrimaryKey) CheckUnsigned() error {
	if !mysql.HasUnsignedFlag(p.Flag) && global.App.AuditConfig.CHECK_PRIMARYKEY_USE_UNSIGNED {
		// bigint必须定义unsigned
		return fmt.Errorf("表`%s`的主键%s必须定义unsigned", p.Table, p.Column)
	}
	return nil
}

func (p *PrimaryKey) CheckAutoIncrement() error {
	if !p.HasAutoIncrement && global.App.AuditConfig.CHECK_PRIMARYKEY_USE_AUTO_INCREMENT {
		// 必须定义AUTO_INCREMENT
		return fmt.Errorf("表`%s`的主键`%s`必须定义auto_increment", p.Table, p.Column)
	}
	return nil
}

func (p *PrimaryKey) CheckNotNull() error {
	if !p.HasNotNull {
		// 必须定义NOT NULL
		return fmt.Errorf("表`%s`的主键%s必须定义NOT NULL", p.Table, p.Column)
	}
	return nil
}
