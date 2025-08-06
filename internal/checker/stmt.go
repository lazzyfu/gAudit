package checker

import (
	"regexp"

	"github.com/lazzyfu/gaudit/internal/config"
	"github.com/lazzyfu/gaudit/internal/dao"
	"github.com/lazzyfu/gaudit/internal/hint"
	"github.com/lazzyfu/gaudit/internal/rules"
	"github.com/lazzyfu/gaudit/pkg/kv"

	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
)

// Stmt结构体
type Stmt struct {
	DB          *dao.DB
	Stmt        ast.StmtNode
	KV          *kv.KVCache[string]
	FingerId    string
	AuditConfig config.AuditConfiguration
}

// 工厂方法
func NewStmt(ctx *StmtContext, stmt ast.StmtNode) Stmt {
	return Stmt{
		DB:          ctx.DB,
		Stmt:        stmt,
		KV:          ctx.Kv,
		FingerId:    ctx.FingerId,
		AuditConfig: ctx.AuditConfig,
	}
}

// 建表语句审核
func (s *Stmt) CreateTableStmt() ReturnData {
	var data ReturnData = ReturnData{FingerId: s.FingerId, Query: s.Stmt.Text(), Type: "CreateTable", Level: "INFO"}

	for _, rule := range rules.CreateTableRules() {
		var ruleHint *hint.RuleHint = &hint.RuleHint{
			DB:          s.DB,
			KV:          s.KV,
			Query:       s.Stmt.Text(),
			AuditConfig: &s.AuditConfig,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &s.Stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// 创建视图语句审核
func (s *Stmt) CreateViewStmt() ReturnData {
	var data ReturnData = ReturnData{FingerId: s.FingerId, Query: s.Stmt.Text(), Type: "CreateView", Level: "INFO"}

	for _, rule := range rules.CreateViewRules() {
		var ruleHint *hint.RuleHint = &hint.RuleHint{
			DB:          s.DB,
			KV:          s.KV,
			Query:       s.Stmt.Text(),
			AuditConfig: &s.AuditConfig,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &s.Stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// RENAME语句审核
func (s *Stmt) RenameTableStmt() ReturnData {
	var data ReturnData = ReturnData{FingerId: s.FingerId, Query: s.Stmt.Text(), Type: "RenameTable", Level: "INFO"}

	for _, rule := range rules.RenameTableRules() {
		var ruleHint *hint.RuleHint = &hint.RuleHint{
			DB:          s.DB,
			KV:          s.KV,
			Query:       s.Stmt.Text(),
			AuditConfig: &s.AuditConfig,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &s.Stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// ANALYZE语句审核
func (s *Stmt) AnalyzeTableStmt() ReturnData {
	var data ReturnData = ReturnData{FingerId: s.FingerId, Query: s.Stmt.Text(), Type: "AnalyzeTable", Level: "INFO"}

	for _, rule := range rules.AnalyzeTableRules() {
		var ruleHint *hint.RuleHint = &hint.RuleHint{
			DB:          s.DB,
			KV:          s.KV,
			Query:       s.Stmt.Text(),
			AuditConfig: &s.AuditConfig,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &s.Stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// DROP/TRUNCATE语句审核
func (s *Stmt) DropTableStmt() ReturnData {
	var data ReturnData = ReturnData{FingerId: s.FingerId, Query: s.Stmt.Text(), Type: "DropTable", Level: "INFO"}

	for _, rule := range rules.DropTableRules() {
		var ruleHint *hint.RuleHint = &hint.RuleHint{
			DB:          s.DB,
			KV:          s.KV,
			Query:       s.Stmt.Text(),
			AuditConfig: &s.AuditConfig,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &s.Stmt)

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}

// ALTER语句审核，返回ReturnData和表名（用于合并提示）
func (s *Stmt) AlterTableStmt() (ReturnData, string) {
	var data ReturnData = ReturnData{FingerId: s.FingerId, Query: s.Stmt.Text(), Type: "AlterTable", Level: "INFO"}
	var mergeAlter string
	// 禁止使用ALTER TABLE...ADD CONSTRAINT...语法
	tmpCompile := regexp.MustCompile(`(?is:.*alter.*table.*add.*constraint.*)`)
	match := tmpCompile.MatchString(s.Stmt.Text())
	if match {
		data.Level = "WARN"
		data.Summary = append(data.Summary, "禁止使用ALTER TABLE...ADD CONSTRAINT...语法")
		return data, mergeAlter
	}

	for _, rule := range rules.AlterTableRules() {
		var ruleHint *hint.RuleHint = &hint.RuleHint{
			DB:          s.DB,
			KV:          s.KV,
			AuditConfig: &s.AuditConfig,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &s.Stmt)
		if len(rule.RuleHint.MergeAlter) > 0 && len(mergeAlter) == 0 {
			mergeAlter = rule.RuleHint.MergeAlter
		}
		if len(rule.RuleHint.Summary) > 0 {
			// 检查不通过
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			// 如果IsSkipNextStep为true，跳过接下来的检查步骤
			break
		}
	}

	return data, mergeAlter
}

// DML语句审核
func (s *Stmt) DMLStmt() ReturnData {
	// delete/update/insert语句
	/*
		DML语句真的需要对同一个指纹的SQL跳过校验？
		1. DML规则并不多，对实际校验性能影响不大
		2. 每条DML都需要进行Explain，由于考虑传值不一样，因此指纹一样并不能代表Explain的影响行数一样
		3. 实际测试1000条update校验仅需800ms,2000条update校验仅需1500ms
	*/
	var data ReturnData = ReturnData{FingerId: s.FingerId, Query: s.Stmt.Text(), Type: "DML", Level: "INFO"}

	for _, rule := range rules.DMLRules() {
		var ruleHint *hint.RuleHint = &hint.RuleHint{
			DB:          s.DB,
			KV:          s.KV,
			Query:       s.Stmt.Text(),
			AuditConfig: &s.AuditConfig,
		}
		rule.RuleHint = ruleHint
		rule.CheckFunc(&rule, &s.Stmt)

		// 当为DML语句时，赋值AffectedRows
		data.AffectedRows = rule.RuleHint.AffectedRows

		if len(rule.RuleHint.Summary) > 0 {
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.RuleHint.Summary...)
		}
		if rule.RuleHint.IsSkipNextStep {
			break
		}
	}

	return data
}
