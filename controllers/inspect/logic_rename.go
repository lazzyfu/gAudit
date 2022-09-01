/*
@Time    :   2022/08/25 16:42:48
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/common/utils"
)

// LogicRenameTable
func LogicRenameTable(v *TraverseRenameTable, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if !r.AuditConfig.ENABLE_RENAME_TABLE_NAME {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许RENAME表名[表`%s`]", v.OldTable))
		return
	}
	// 禁止审核指定的表
	if len(r.AuditConfig.DISABLE_AUDIT_DDL_TABLES) > 0 {
		for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
			if item.DB == r.DB.Database && utils.IsContain(item.Tables, v.OldTable) {
				r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核,原因: %s", r.DB.Database, v.OldTable, item.Reason))
			}
		}
	}
	// 旧表必须存在
	if err, msg := DescTable(v.OldTable, r.DB); err != nil {
		r.Summary = append(r.Summary, msg)
	}
	// 新表不能存在
	if err, msg := DescTable(v.NewTable, r.DB); err == nil {
		r.Summary = append(r.Summary, msg)
	}
}
