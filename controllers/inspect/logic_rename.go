/*
@Time    :   2022/08/25 16:42:48
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
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
	// 旧表必须存在
	if err, msg := DescTable(v.OldTable, r.DB); err != nil {
		r.Summary = append(r.Summary, msg)
	}
	// 新表不能存在
	if err, msg := DescTable(v.NewTable, r.DB); err == nil {
		r.Summary = append(r.Summary, msg)
	}
}
