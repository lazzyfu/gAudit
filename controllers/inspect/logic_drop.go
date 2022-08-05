/*
@Time    :   2022/07/06 10:12:27
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/global"
)

// LogicDropTable
func LogicDropTable(v *TraverseDropTable, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.IsHasDropTable {
		if !global.App.AuditConfig.ENABLE_DROP_TABLE {
			r.Summary = append(r.Summary, fmt.Sprintf("禁止DROP[表%s]", v.Tables))
			return
		}
		// 检查表是否存在
		for _, table := range v.Tables {
			if err, msg := DescTable(table, r.DB); err != nil {
				r.Summary = append(r.Summary, msg)
			}
		}
	}
}

// LogicTruncateTable
func LogicTruncateTable(v *TraverseTruncateTable, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.IsHasTruncateTable {
		if !global.App.AuditConfig.ENABLE_TRUNCATE_TABLE {
			r.Summary = append(r.Summary, fmt.Sprintf("禁止TRUNCATE[表%s]", v.Table))
			return
		}
		// 检查表是否存在
		if err, msg := DescTable(v.Table, r.DB); err != nil {
			r.Summary = append(r.Summary, msg)
		}
	}
}
