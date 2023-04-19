/*
@Time    :   2023/04/19 15:09:38
@Author  :   zongfei.fu
@Desc    :
*/

package inspect

import "sqlSyntaxAudit/controllers/process"

// LogicRenameTable
func LogicAnalyzeTable(v *TraverseAnalyzeTable, r *Rule) {
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
		if err, msg := DescTable(table, r.DB); err != nil {
			r.Summary = append(r.Summary, msg)
		}
	}
}
