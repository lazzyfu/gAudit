package logics

import (
	"fmt"

	"github.com/lazzyfu/gaudit/internal/config"
	"github.com/lazzyfu/gaudit/internal/hint"
	"github.com/lazzyfu/gaudit/pkg/utils"

	"github.com/lazzyfu/gaudit/internal/dao"
	"github.com/lazzyfu/gaudit/internal/process"
	"github.com/lazzyfu/gaudit/internal/traverses"
)

// LogicDisableAuditDMLTables
func LogicDisableAuditDMLTables(v *traverses.TraverseDisableAuditDMLTables, r *hint.RuleHint) {
	// 禁止审核指定的表
	if len(r.AuditConfig.DISABLE_AUDIT_DML_TABLES) > 0 {
		for _, item := range r.AuditConfig.DISABLE_AUDIT_DML_TABLES {
			for _, table := range v.Tables {
				if item.DB == r.DB.Database && utils.IsContain(item.Tables, table) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DML语法审核，原因: %s", r.DB.Database, table, item.Reason))
					r.IsSkipNextStep = true
				}
			}
		}
	}
	// DML语句检查表是否存在
	for _, table := range v.Tables {
		if err, msg := dao.DescTable(table, r.DB); err != nil {
			r.Summary = append(r.Summary, msg)
			r.IsSkipNextStep = true
		}
	}
}

// LogicDMLInsertIntoSelect
func LogicDMLInsertIntoSelect(v *traverses.TraverseDMLInsertIntoSelect, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	if r.AuditConfig.DISABLE_INSERT_INTO_SELECT && v.HasSelectSubQuery {
		r.Summary = append(r.Summary, fmt.Sprintf("禁止使用%s into select语法", v.DMLType))
		r.IsSkipNextStep = true
	}
	if r.AuditConfig.DISABLE_ON_DUPLICATE && v.HasOnDuplicate {
		r.Summary = append(r.Summary, fmt.Sprintf("禁止使用%s into on duplicate语法", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLNoWhere
func LogicDMLNoWhere(v *traverses.TraverseDMLNoWhere, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	if !v.HasWhere && r.AuditConfig.DML_MUST_HAVE_WHERE {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句必须要有where条件", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLInsertWithColumns
func LogicDMLInsertWithColumns(v *traverses.TraverseDMLInsertWithColumns, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	if v.DMLType == "REPLACE" && r.AuditConfig.DISABLE_REPLACE {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许使用%s语句", v.DMLType))
		r.IsSkipNextStep = true
		return
	}
	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	// 判断列是否存在
	for _, col := range v.Columns {
		if !utils.IsContain(vAudit.Cols, col) {
			r.Summary = append(r.Summary, fmt.Sprintf("列`%s`不存在[表`%s`]", col, v.Table))
		}
	}
	// 强制指定列名
	if v.ColumnsCount == 0 {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句必须指定列名", v.DMLType))
	} else if !v.ColsValuesIsMatch {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句指定的列数量和值的数量不匹配", v.DMLType))
	}
	if v.RowsCount > r.AuditConfig.MAX_INSERT_ROWS {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句单次最多允许的行数为%d，当前行数为%d【建议拆分为多条%s语句】", v.DMLType, r.AuditConfig.MAX_INSERT_ROWS, v.RowsCount, v.DMLType))
	}
}

// LogicDMLHasLimit
func LogicDMLHasConstraint(v *traverses.TraverseDMLHasConstraint, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	if v.HasLimit && r.AuditConfig.DML_DISABLE_LIMIT {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句不能有LIMIT子句", v.DMLType))
		r.IsSkipNextStep = true
	}
	if v.HasOrderBy && r.AuditConfig.DML_DISABLE_ORDERBY {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句不能有ORDER BY子句", v.DMLType))
		r.IsSkipNextStep = true
	}
	if v.HasSubQuery && r.AuditConfig.DML_DISABLE_SUBQUERY {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句不能有子查询", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLJoinWithOn
func LogicDMLJoinWithOn(v *traverses.TraverseDMLJoinWithOn, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	if v.HasJoin && r.AuditConfig.CHECK_DML_JOIN_WITH_ON && !v.IsJoinWithOn {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句的JOIN操作必须要有ON条件", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLMaxUpdateRows
func LogicDMLMaxUpdateRows(v *traverses.TraverseDMLMaxUpdateRows, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	explain := process.Explain{DB: r.DB, SQL: r.Query, KV: r.KV}
	affectedRows, err := explain.Get(r.AuditConfig.EXPLAIN_RULE)
	if err != nil {
		r.AffectedRows = 0
		r.Summary = append(r.Summary, err.Error())
		r.IsSkipNextStep = true
		return
	}
	if affectedRows > r.AuditConfig.MAX_AFFECTED_ROWS {
		r.AffectedRows = affectedRows
		r.Summary = append(r.Summary, fmt.Sprintf("当前%s语句影响或扫描行数超过%d，建议拆分为多条语句，确保每条语句影响或扫描行数不超过%d", v.DMLType, r.AuditConfig.MAX_AFFECTED_ROWS, r.AuditConfig.MAX_AFFECTED_ROWS))
		r.IsSkipNextStep = true
		return
	}
	r.IsSkipNextStep = true
	r.AffectedRows = affectedRows
}

// LogicDMLMaxInsertRows
func LogicDMLMaxInsertRows(v *traverses.TraverseDMLMaxInsertRows, r *hint.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.AffectedRows = v.RowsCount
	r.IsSkipNextStep = true
}
