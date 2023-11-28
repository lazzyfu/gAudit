/*
@Time    :   2022/07/06 10:12:20
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/controllers/process"
)

// LogicDisableAuditDMLTables
func LogicDisableAuditDMLTables(v *TraverseDisableAuditDMLTables, r *Rule) {
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
		if err, msg := DescTable(table, r.DB); err != nil {
			r.Summary = append(r.Summary, msg)
			r.IsSkipNextStep = true
		}
	}
}

// LogicDMLInsertIntoSelect
func LogicDMLInsertIntoSelect(v *TraverseDMLInsertIntoSelect, r *Rule) {
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
func LogicDMLNoWhere(v *TraverseDMLNoWhere, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if !v.HasWhere && r.AuditConfig.DML_MUST_HAVE_WHERE {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句必须要有where条件", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLInsertWithColumns
func LogicDMLInsertWithColumns(v *TraverseDMLInsertWithColumns, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.DMLType == "REPLACE" && r.AuditConfig.DISABLE_REPLACE {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许使用%s语句", v.DMLType))
		r.IsSkipNextStep = true
		return
	}
	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &TraverseAlterTableShowCreateTableGetCols{}
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
func LogicDMLHasConstraint(v *TraverseDMLHasConstraint, r *Rule) {
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
func LogicDMLJoinWithOn(v *TraverseDMLJoinWithOn, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.HasJoin && r.AuditConfig.CHECK_DML_JOIN_WITH_ON && !v.IsJoinWithOn {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句的JOIN操作必须要有ON条件", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLMaxUpdateRows
func LogicDMLMaxUpdateRows(v *TraverseDMLMaxUpdateRows, r *Rule) {
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
		r.Summary = append(r.Summary, fmt.Sprintf("当前%s语句最大影响或扫描行数超过了最大允许值%d【建议您将语句拆分为多条，保证每条语句影响或扫描行数小于最大允许值%d】", v.DMLType, r.AuditConfig.MAX_AFFECTED_ROWS, r.AuditConfig.MAX_AFFECTED_ROWS))
		r.IsSkipNextStep = true
		return
	}
	r.IsSkipNextStep = true
	r.AffectedRows = affectedRows
}

// LogicDMLMaxInsertRows
func LogicDMLMaxInsertRows(v *TraverseDMLMaxInsertRows, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.AffectedRows = v.RowsCount
	r.IsSkipNextStep = true
}
