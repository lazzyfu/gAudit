/*
@Time    :   2022/07/06 10:12:42
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"github.com/pingcap/tidb/parser/ast"
)

func CreateTableRules() []Rule {
	return []Rule{
		{
			Hint:      "CreateTable#检查表是否存在",
			CheckFunc: (*Rule).RuleCreateTableIsExist,
		},
		{
			Hint:      "CreateTable#检查CreateTableAs语法",
			CheckFunc: (*Rule).RuleCreateTableAs,
		},
		{
			Hint:      "CreateTable#检查CreateTableLike语法",
			CheckFunc: (*Rule).RuleCreateTableLike,
		},
		{
			Hint:      "CreateTable#表Options检查",
			CheckFunc: (*Rule).RuleCreateTableOptions,
		},
		{
			Hint:      "CreateTable#主键检查",
			CheckFunc: (*Rule).RuleCreateTablePrimaryKey,
		},
		{
			Hint:      "CreateTable#约束检查",
			CheckFunc: (*Rule).RuleCreateTableConstraint,
		},
		{
			Hint:      "CreateTable#审计字段检查",
			CheckFunc: (*Rule).RuleCreateTableAuditCols,
		},
		{
			Hint:      "CreateTable#列Options检查",
			CheckFunc: (*Rule).RuleCreateTableColsOptions,
		},
		{
			Hint:      "CreateTable#列重复定义检查",
			CheckFunc: (*Rule).RuleCreateTableColsRepeatDefine,
		},
		{

			Hint:      "CreateTable#列字符集检查",
			CheckFunc: (*Rule).RuleCreateTableColsCharset,
		},
		{
			Hint: "CreateTable#索引前缀检查",

			CheckFunc: (*Rule).RuleCreateTableIndexesPrefix,
		},
		{
			Hint:      "CreateTable#索引数量检查",
			CheckFunc: (*Rule).RuleCreateTableIndexesCount,
		},
		{
			Hint:      "CreateTable#索引重复定义检查",
			CheckFunc: (*Rule).RuleCreateTableIndexesRepeatDefine,
		},
		{
			Hint:      "CreateTable#冗余索引检查",
			CheckFunc: (*Rule).RuleCreateTableRedundantIndexes,
		},
		{
			Hint:      "CreateTable#BLOB/TEXT类型不能设置为索引",
			CheckFunc: (*Rule).RuleCreateTableDisabledIndexes,
		},
		{
			Hint:      "CreateTable#索引InnodbLargePrefix",
			CheckFunc: (*Rule).RuleCreateTableInnodbLargePrefix,
		},
		{
			Hint:      "CreateTable#检查表定义的行是否超过65535",
			CheckFunc: (*Rule).RuleCreateTableInnoDBRowSize,
		},
	}
}

// RuleCreateTableIsExist
func (r *Rule) RuleCreateTableIsExist(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableIsExist{}
	(*tistmt).Accept(v)
	LogicCreateTableIsExist(v, r)
}

// RuleCreateTableAs
func (r *Rule) RuleCreateTableAs(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableAs{}
	(*tistmt).Accept(v)
	LogicCreateTableAs(v, r)
}

// RuleCreateTableLike
func (r *Rule) RuleCreateTableLike(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableLike{}
	(*tistmt).Accept(v)
	LogicCreateTableLike(v, r)
}

// RuleCreateTableOptions
func (r *Rule) RuleCreateTableOptions(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableOptions{}
	(*tistmt).Accept(v)
	LogicCreateTableOptions(v, r)
}

// RuleCreateTablePrimaryKey
func (r *Rule) RuleCreateTablePrimaryKey(tistmt *ast.StmtNode) {
	v := &TraverseCreateTablePrimaryKey{}
	(*tistmt).Accept(v)
	LogicCreateTablePrimaryKey(v, r)
}

// RuleCreateTableConstraint
func (r *Rule) RuleCreateTableConstraint(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableConstraint{}
	(*tistmt).Accept(v)
	LogicCreateTableConstraint(v, r)
}

// RuleCreateTableAuditCols
func (r *Rule) RuleCreateTableAuditCols(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableAuditCols{}
	(*tistmt).Accept(v)
	LogicCreateTableAuditCols(v, r)
}

// RuleCreateTableColsOptions
func (r *Rule) RuleCreateTableColsOptions(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableColsOptions{}
	(*tistmt).Accept(v)
	LogicCreateTableColsOptions(v, r)
}

// RuleCreateTableColsRepeatDefine
func (r *Rule) RuleCreateTableColsRepeatDefine(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableColsRepeatDefine{}
	(*tistmt).Accept(v)
	LogicCreateTableColsRepeatDefine(v, r)
}

// RuleCreateTableColsCharset
func (r *Rule) RuleCreateTableColsCharset(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableColsCharset{}
	(*tistmt).Accept(v)
	LogicCreateTableColsCharset(v, r)
}

// RuleCreateTableIndexesPrefix
func (r *Rule) RuleCreateTableIndexesPrefix(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableIndexesPrefix{}
	(*tistmt).Accept(v)
	LogicCreateTableIndexesPrefix(v, r)
}

// RuleCreateTableIndexesCount
func (r *Rule) RuleCreateTableIndexesCount(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableIndexesCount{}
	(*tistmt).Accept(v)
	LogicCreateTableIndexesCount(v, r)
}

// RuleCreateTableIndexesRepeatDefine
func (r *Rule) RuleCreateTableIndexesRepeatDefine(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableIndexesRepeatDefine{}
	(*tistmt).Accept(v)
	LogicCreateTableIndexesRepeatDefine(v, r)
}

// RuleCreateTableRedundantIndexes
func (r *Rule) RuleCreateTableRedundantIndexes(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableRedundantIndexes{}
	(*tistmt).Accept(v)
	LogicCreateTableRedundantIndexes(v, r)
}

// RuleCreateTableDisabledIndexes
func (r *Rule) RuleCreateTableDisabledIndexes(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableDisabledIndexes{}
	(*tistmt).Accept(v)
	LogicCreateTableDisabledIndexes(v, r)
}

// RuleCreateTableInnodbLargePrefix
func (r *Rule) RuleCreateTableInnodbLargePrefix(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableInnodbLargePrefix{}
	(*tistmt).Accept(v)
	LogicCreateTableInnodbLargePrefix(v, r)
}

// RuleCreateTableInnoDBRowSize
func (r *Rule) RuleCreateTableInnoDBRowSize(tistmt *ast.StmtNode) {
	v := &TraverseCreateTableInnoDBRowSize{}
	(*tistmt).Accept(v)
	LogicCreateTableInnoDBRowSize(v, r)
}
