/*
@Time    :   2023/04/19 15:09:38
@Author  :   xff
@Desc    :
*/

package logics

import (
	"gAudit/controllers"
	"gAudit/controllers/dao"
	"gAudit/controllers/process"
	"gAudit/controllers/traverses"
)

// LogicRenameTable
func LogicAnalyzeTable(v *traverses.TraverseAnalyzeTable, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	dbVersionIns := process.DbVersion{Version: r.KV.Get("dbVersion").(string)}
	if !dbVersionIns.IsTiDB() {
		r.Summary = append(r.Summary, "仅允许TiDB提交Analyze table语法")
		return
	}
	// 表必须存在
	for _, table := range v.TableNames {
		if err, msg := dao.DescTable(table, r.DB); err != nil {
			r.Summary = append(r.Summary, msg)
		}
	}
}
