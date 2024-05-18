/*
@Time    :   2022/07/06 10:12:27
@Author  :   xff
@Desc    :   None
*/

package logics

import (
	"fmt"
	"gAudit/controllers"
	"gAudit/controllers/dao"
	"gAudit/controllers/process"
	"gAudit/controllers/traverses"
	"gAudit/pkg/utils"
)

func CheckTableDropAndTruncate(table string, r *controllers.RuleHint) {
	if err, msg := dao.DescTable(table, r.DB); err != nil {
		r.Summary = append(r.Summary, msg)
		return
	}

	dbVersion := process.DbVersion{Version: r.KV.Get("dbVersion").(string)}
	innodbAdaptiveHashIndex := r.KV.Get("innodbAdaptiveHashIndex").(string)
	if !dbVersion.IsTiDB() && dbVersion.Int() < 80023 && innodbAdaptiveHashIndex == "ON" {
		if err := dao.CheckTableRowCountLimit(table, r.AuditConfig.DT_TABLE_MAXROW_LIMIT, r.DB); err != nil {
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`存在数据，执行DROP/TRUNCATE操作存在风险(自适应哈希索引清理可能阻塞其他语句执行)，请先联系DBA完成数据清理", table))
		}
	}
}

// LogicDropTable
func LogicDropTable(v *traverses.TraverseDropTable, r *controllers.RuleHint) {
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
func LogicTruncateTable(v *traverses.TraverseTruncateTable, r *controllers.RuleHint) {
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
