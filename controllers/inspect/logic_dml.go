/*
@Time    :   2022/07/06 10:12:20
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/controllers/process"
	"sqlSyntaxAudit/global"
)

// LogicDMLInsertIntoSelect
func LogicDMLInsertIntoSelect(v *TraverseDMLInsertIntoSelect, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if global.App.AuditConfig.DISABLE_INSERT_INTO_SELECT && v.HasSelectSubQuery {
		r.Summary = append(r.Summary, fmt.Sprintf("禁止使用%s into select语法", v.DMLType))
		r.IsSkipNextStep = true
	}
	if global.App.AuditConfig.DISABLE_ON_DUPLICATE && v.HasOnDuplicate {
		r.Summary = append(r.Summary, fmt.Sprintf("禁止使用%s into on duplicate语法", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLNoWhere
func LogicDMLNoWhere(v *TraverseDMLNoWhere, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if !v.HasWhere && global.App.AuditConfig.DML_MUST_HAVE_WHERE {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句必须要有where条件", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLInsertWithColumns
func LogicDMLInsertWithColumns(v *TraverseDMLInsertWithColumns, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.DMLType == "REPLACE" && global.App.AuditConfig.DISABLE_REPLACE {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许使用%s语句", v.DMLType))
		r.IsSkipNextStep = true
		return
	}
	// 强制指定列名
	if v.ColumnsCount == 0 {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句必须指定列名", v.DMLType))
	} else if !v.ColsValuesIsMatch {
		// todo 检查列名是否存在
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句指定的列数量和值的数量不匹配", v.DMLType))
	}
	if v.RowsCount > global.App.AuditConfig.MAX_INSERT_ROWS {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句单次最多允许的行数为%d,当前行数为%d「可以拆分为多条%s语句」", v.DMLType, global.App.AuditConfig.MAX_INSERT_ROWS, v.RowsCount, v.DMLType))
	}
}

// LogicDMLHasLimit
func LogicDMLHasConstraint(v *TraverseDMLHasConstraint, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.HasLimit && global.App.AuditConfig.DML_DISABLE_LIMIT {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句不能有LIMIT子句", v.DMLType))
		r.IsSkipNextStep = true
	}
	if v.HasOrderBy && global.App.AuditConfig.DML_DISABLE_ORDERBY {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句不能有ORDER BY子句", v.DMLType))
		r.IsSkipNextStep = true
	}
	if v.HasSubQuery && global.App.AuditConfig.DML_DISABLE_SUBQUERY {
		r.Summary = append(r.Summary, fmt.Sprintf("%s语句不能有子查询", v.DMLType))
		r.IsSkipNextStep = true
	}
}

// LogicDMLJoinWithOn
func LogicDMLJoinWithOn(v *TraverseDMLJoinWithOn, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.HasJoin && global.App.AuditConfig.CHECK_DML_JOIN_WITH_ON && !v.IsJoinWithOn {
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
	affectedRows, err := explain.Get()
	if err != nil {
		r.AffectedRows = 0
		r.Summary = append(r.Summary, err.Error())
		return
	}
	if affectedRows > global.App.AuditConfig.MAX_AFFECTED_ROWS {
		r.AffectedRows = affectedRows
		r.Summary = append(r.Summary, fmt.Sprintf("当前%s语句最大影响行数超过了最大允许值%d", v.DMLType, global.App.AuditConfig.MAX_AFFECTED_ROWS))
		return
	}
	r.AffectedRows = affectedRows
}
