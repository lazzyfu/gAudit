package logics

import (
	"fmt"

	"github.com/lazzyfu/gaudit/pkg/utils"

	"github.com/lazzyfu/gaudit/internal/dao"
	"github.com/lazzyfu/gaudit/internal/hint"
	"github.com/lazzyfu/gaudit/internal/process"
	"github.com/lazzyfu/gaudit/internal/traverses"
)

func CheckTableDropAndTruncate(table string, r *hint.RuleHint) {
	if err, msg := dao.DescTable(table, r.DB); err != nil {
		r.Summary = append(r.Summary, msg)
		return
	}
	dbVersion, ok := r.KV.Get("dbVersion")
	if !ok {
		r.Summary = append(r.Summary, "无法获取数据库版本信息")
		return
	}
	innodbAdaptiveHashIndex, ok := r.KV.Get("innodbAdaptiveHashIndex")
	if !ok {
		r.Summary = append(r.Summary, "无法获取innodb_adaptive_hash_index参数")
		return
	}
	dbVersionIns := process.DbVersion{Version: dbVersion}
	if !dbVersionIns.IsTiDB() && dbVersionIns.Int() < 80023 && innodbAdaptiveHashIndex == "ON" {
		if err := dao.CheckTableRowCountLimit(table, r.AuditConfig.DT_TABLE_MAXROW_LIMIT, r.DB); err != nil {
			r.Summary = append(r.Summary, fmt.Sprintf("当前数据库版本`%s`，执行DROP/TRUNCATE操作存在潜在阻塞风险（遍历缓冲池并驱逐要删除的页），建议先清理表`%s`数据", dbVersion, table))
		}
	}
}

// LogicDropTable
func LogicDropTable(v *traverses.TraverseDropTable, r *hint.RuleHint) {
	if !v.IsMatch {
		return
	}
	if !r.AuditConfig.ENABLE_DROP_TABLE {
		r.Summary = append(r.Summary, fmt.Sprintf("禁止DROP表：%s", v.Tables))
		return
	}
	// 禁止审核指定的表
	for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
		for _, table := range v.Tables {
			if item.DB == r.DB.Database && utils.IsContain(item.Tables, table) {
				r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核，原因: %s", r.DB.Database, table, item.Reason))
			}
		}
	}

	// 检查表的DROP操作
	for _, table := range v.Tables {
		CheckTableDropAndTruncate(table, r)
	}
}

// LogicTruncateTable
func LogicTruncateTable(v *traverses.TraverseTruncateTable, r *hint.RuleHint) {
	if !v.IsMatch {
		return
	}
	if !r.AuditConfig.ENABLE_TRUNCATE_TABLE {
		r.Summary = append(r.Summary, fmt.Sprintf("禁止TRUNCATE表：`%s`", v.Table))
		return
	}
	// 禁止审核指定的表
	for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
		if item.DB == r.DB.Database && utils.IsContain(item.Tables, v.Table) {
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核，原因: %s", r.DB.Database, v.Table, item.Reason))
		}
	}
	CheckTableDropAndTruncate(v.Table, r)
}
