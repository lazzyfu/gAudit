/*
@Time    :   2022/06/28 10:21:41
@Author  :   zongfei.fu
@Desc    :   alter规则逻辑，Level初始化为INFO
*/

package inspect

import (
	"github.com/pingcap/tidb/parser/ast"
)

func AlterTableRules() []Rule {
	return []Rule{
		{
			Hint:      "AlterTable#检查表是否存在",
			CheckFunc: (*Rule).RuleAlterTableIsExist,
		},
		{
			Hint:      "AlterTable#检查TiDBMergeAlter",
			CheckFunc: (*Rule).RuleAlterTiDBMerge,
		},
		{
			Hint:      "AlterTable#DROP列和索引检查",
			CheckFunc: (*Rule).RuleAlterTableDropColsOrIndexes,
		},
		{
			Hint:      "AlterTable#DropTiDBColWithCoveredIndex检查",
			CheckFunc: (*Rule).RuleAlterTableDropTiDBColWithCoveredIndex,
		},
		{
			Hint:      "AlterTable#表Options检查",
			CheckFunc: (*Rule).RuleAlterTableOptions,
		},
		{
			Hint:      "AlterTable#列字符集检查",
			CheckFunc: (*Rule).RuleAlterTableColCharset,
		},
		{
			Hint:      "AlterTable#Add列Options检查",
			CheckFunc: (*Rule).RuleAlterTableAddColOptions,
		},
		{
			Hint:      "AlterTable#Add主键检查",
			CheckFunc: (*Rule).RuleAlterTableAddPrimaryKey,
		},
		{
			Hint:      "AlterTable#Add重复列检查",
			CheckFunc: (*Rule).RuleAlterTableAddColRepeatDefine,
		},
		{
			Hint:      "AlterTable#Add索引前缀检查",
			CheckFunc: (*Rule).RuleAlterTableAddIndexPrefix,
		},
		{
			Hint:      "AlterTable#Add索引数量检查",
			CheckFunc: (*Rule).RuleAlterTableAddIndexCount,
		},
		{
			Hint:      "AlterTable#AddConstraint检查",
			CheckFunc: (*Rule).RuleAlterTableAddConstraint,
		},
		{
			Hint:      "AlterTable#Add重复索引检查",
			CheckFunc: (*Rule).RuleAlterTableAddIndexRepeatDefine,
		},
		{
			Hint:      "AlterTable#Add冗余索引检查",
			CheckFunc: (*Rule).RuleAlterTableRedundantIndexes,
		},
		{
			Hint:      "AlterTable#BLOB/TEXT类型不能设置为索引",
			CheckFunc: (*Rule).RuleAlterTableDisabledIndexes,
		},
		{
			Hint:      "AlterTable#Modify列Options检查",
			CheckFunc: (*Rule).RuleAlterTableModifyColOptions,
		},
		{
			Hint:      "AlterTable#Change列Options检查",
			CheckFunc: (*Rule).RuleAlterTableChangeColOptions,
		},
		{
			Hint:      "AlterTable#RenameIndex检查",
			CheckFunc: (*Rule).RuleAlterTableRenameIndex,
		},
		{
			Hint:      "AlterTable#RenameTblName检查",
			CheckFunc: (*Rule).RuleAlterTableRenameTblName,
		},
		{
			Hint:      "AlterTable#索引InnodbLargePrefix",
			CheckFunc: (*Rule).RuleAlterTableInnodbLargePrefix,
		},
		{
			Hint:      "AlterTable#检查表定义的行是否超过65535",
			CheckFunc: (*Rule).RuleAlterTableRowSizeTooLarge,
		},
	}
}

// RuleAlterTableIsExist
func (r *Rule) RuleAlterTableIsExist(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableIsExist{}
	(*tistmt).Accept(v)
	LogicAlterTableIsExist(v, r)
}

// RuleAlterTiDBMerge
func (r *Rule) RuleAlterTiDBMerge(tistmt *ast.StmtNode) {
	v := &TraverseAlterTiDBMerge{}
	(*tistmt).Accept(v)
	LogicAlterTableTiDBMerge(v, r)
}

// RuleAlterTableDropCols
func (r *Rule) RuleAlterTableDropColsOrIndexes(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableDropColsOrIndexes{}
	(*tistmt).Accept(v)
	LogicAlterTableDropColsOrIndexes(v, r)
}

// RuleAlterTableDropTiDBColWithCoveredIndex
func (r *Rule) RuleAlterTableDropTiDBColWithCoveredIndex(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableDropTiDBColWithCoveredIndex{}
	(*tistmt).Accept(v)
	LogicAlterTableDropTiDBColWithCoveredIndex(v, r)
}

// RuleAlterTableOptions
func (r *Rule) RuleAlterTableOptions(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableOptions{}
	(*tistmt).Accept(v)
	LogicAlterTableOptions(v, r)
}

// RuleAlterTableColCharset
func (r *Rule) RuleAlterTableColCharset(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableColCharset{}
	(*tistmt).Accept(v)
	LogicAlterTableColCharset(v, r)
}

// RuleAlterTableAddColOptions
func (r *Rule) RuleAlterTableAddColOptions(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableAddColOptions{}
	(*tistmt).Accept(v)
	LogicAlterTableAddColOptions(v, r)
}

// RuleAlterTableAddColWithPrimaryKey
func (r *Rule) RuleAlterTableAddPrimaryKey(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableAddPrimaryKey{}
	(*tistmt).Accept(v)
	LogicAlterTableAddPrimaryKey(v, r)
}

// RuleAlterTableAddColRepeatDefine
func (r *Rule) RuleAlterTableAddColRepeatDefine(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableAddColRepeatDefine{}
	(*tistmt).Accept(v)
	LogicAlterTableAddColRepeatDefine(v, r)
}

// RuleAlterTableAddIndexPrefix
func (r *Rule) RuleAlterTableAddIndexPrefix(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableAddIndexPrefix{}
	(*tistmt).Accept(v)
	LogicAlterTableAddIndexPrefix(v, r)
}

// RuleAlterTableAddIndexCount
func (r *Rule) RuleAlterTableAddIndexCount(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableAddIndexCount{}
	(*tistmt).Accept(v)
	LogicAlterTableAddIndexCount(v, r)
}

// RuleAlterTableAddConstraint
func (r *Rule) RuleAlterTableAddConstraint(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableAddConstraint{}
	(*tistmt).Accept(v)
	LogicAlterTableAddConstraint(v, r)
}

// RuleAlterTableAddIndexRepeatDefine
func (r *Rule) RuleAlterTableAddIndexRepeatDefine(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableAddIndexRepeatDefine{}
	(*tistmt).Accept(v)
	LogicAlterTableAddIndexRepeatDefine(v, r)
}

// RuleAlterTableRedundantIndexes
func (r *Rule) RuleAlterTableRedundantIndexes(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableRedundantIndexes{}
	(*tistmt).Accept(v)
	LogicAlterTableRedundantIndexes(v, r)
}

// RuleAlterTableDisabledIndexes
func (r *Rule) RuleAlterTableDisabledIndexes(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableDisabledIndexes{}
	(*tistmt).Accept(v)
	LogicAlterTableDisabledIndexes(v, r)
}

// RuleAlterTableModifyColOptions
func (r *Rule) RuleAlterTableModifyColOptions(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableModifyColOptions{}
	(*tistmt).Accept(v)
	LogicAlterTableModifyColOptions(v, r)
}

// RuleAlterTableChangeColOptions
func (r *Rule) RuleAlterTableChangeColOptions(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableChangeColOptions{}
	(*tistmt).Accept(v)
	LogicAlterTableChangeColOptions(v, r)
}

// RuleAlterTableRenameIndex
func (r *Rule) RuleAlterTableRenameIndex(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableRenameIndex{}
	(*tistmt).Accept(v)
	LogicAlterTableRenameIndex(v, r)
}

// RuleAlterTableRenameTblName
func (r *Rule) RuleAlterTableRenameTblName(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableRenameTblName{}
	(*tistmt).Accept(v)
	LogicAlterTableRenameTblName(v, r)
}

// RuleAlterTableInnodbLargePrefix
func (r *Rule) RuleAlterTableInnodbLargePrefix(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableInnodbLargePrefix{}
	(*tistmt).Accept(v)
	LogicAlterTableInnodbLargePrefix(v, r)
}

// RuleAlterTableRowSizeTooLarge
func (r *Rule) RuleAlterTableRowSizeTooLarge(tistmt *ast.StmtNode) {
	v := &TraverseAlterTableRowSizeTooLarge{}
	(*tistmt).Accept(v)
	LogicAlterTableRowSizeTooLarge(v, r)
}
