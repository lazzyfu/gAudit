/*
@Time    :   2022/07/06 10:12:48
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"github.com/pingcap/tidb/parser/ast"
)

func DMLRules() []Rule {
	return []Rule{
		{
			Hint:      "DML#限制部分表进行语法审核",
			CheckFunc: (*Rule).RuleDisableAuditDMLTables,
		},
		{
			Hint:      "DML#是否允许INSERT INTO SELECT语法",
			CheckFunc: (*Rule).RuleDMLInsertIntoSelect,
		},
		{
			Hint:      "DML#必须要有WHERE条件",
			CheckFunc: (*Rule).RuleDMLNoWhere,
		},
		{
			Hint:      "DML#INSERT必须指定列名",
			CheckFunc: (*Rule).RuleDMLInsertWithColumns,
		},
		{
			Hint:      "DML#不能有LIMIT/ORDERBY/SubQuery",
			CheckFunc: (*Rule).RuleDMLHasConstraint,
		},
		{
			Hint:      "DML#JOIN操作必须要有ON语句",
			CheckFunc: (*Rule).RuleDMLJoinWithOn,
		},
		{
			Hint:      "DML#更新影响行数",
			CheckFunc: (*Rule).RuleDMLMaxUpdateRows,
		},
		{
			Hint:      "DML#插入影响行数",
			CheckFunc: (*Rule).RuleDMLMaxInsertRows,
		},
	}
}

// RuleDisableAuditDMLTables
func (r *Rule) RuleDisableAuditDMLTables(tistmt *ast.StmtNode) {
	v := &TraverseDisableAuditDMLTables{}
	(*tistmt).Accept(v)
	LogicDisableAuditDMLTables(v, r)
}

// RuleDMLInsertIntoSelect
func (r *Rule) RuleDMLInsertIntoSelect(tistmt *ast.StmtNode) {
	v := &TraverseDMLInsertIntoSelect{}
	(*tistmt).Accept(v)
	LogicDMLInsertIntoSelect(v, r)
}

// RuleDMLNoWhere
func (r *Rule) RuleDMLNoWhere(tistmt *ast.StmtNode) {
	v := &TraverseDMLNoWhere{}
	(*tistmt).Accept(v)
	LogicDMLNoWhere(v, r)
}

// RuleDMLInsertWithColumns
func (r *Rule) RuleDMLInsertWithColumns(tistmt *ast.StmtNode) {
	v := &TraverseDMLInsertWithColumns{}
	(*tistmt).Accept(v)
	LogicDMLInsertWithColumns(v, r)
}

// RuleDMLHasConstraint
func (r *Rule) RuleDMLHasConstraint(tistmt *ast.StmtNode) {
	v := &TraverseDMLHasConstraint{}
	(*tistmt).Accept(v)
	LogicDMLHasConstraint(v, r)
}

// RuleDMLJoinWithOn
func (r *Rule) RuleDMLJoinWithOn(tistmt *ast.StmtNode) {
	v := &TraverseDMLJoinWithOn{}
	(*tistmt).Accept(v)
	LogicDMLJoinWithOn(v, r)
}

// RuleDMLMaxUpdateRows
func (r *Rule) RuleDMLMaxUpdateRows(tistmt *ast.StmtNode) {
	v := &TraverseDMLMaxUpdateRows{}
	(*tistmt).Accept(v)
	LogicDMLMaxUpdateRows(v, r)
}

// RuleDMLMaxInsertRows
func (r *Rule) RuleDMLMaxInsertRows(tistmt *ast.StmtNode) {
	v := &TraverseDMLMaxInsertRows{}
	(*tistmt).Accept(v)
	LogicDMLMaxInsertRows(v, r)
}
