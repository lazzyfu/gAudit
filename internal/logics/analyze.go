package logics

import (
	"github.com/lazzyfu/gaudit/internal/dao"
	"github.com/lazzyfu/gaudit/internal/hint"
	"github.com/lazzyfu/gaudit/internal/process"
	"github.com/lazzyfu/gaudit/internal/traverses"
)

// LogicRenameTable
func LogicAnalyzeTable(v *traverses.TraverseAnalyzeTable, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	dbVersion, ok := r.KV.Get("dbVersion")
	if !ok {
		r.Summary = append(r.Summary, "无法获取数据库版本信息")
		return
	}
	dbVersionIns := process.DbVersion{Version: dbVersion}
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
