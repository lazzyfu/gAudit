package rules

import (
	"github.com/lazzyfu/gaudit/internal/logics"
	"github.com/lazzyfu/gaudit/internal/traverses"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

func RenameTableRules() []Rule {
	return []Rule{
		{
			Hint:      "RenameTable#检查",
			CheckFunc: (*Rule).RuleRenameTable,
		},
	}
}

// RuleRenameTable
func (r *Rule) RuleRenameTable(tistmt *ast.StmtNode) {
	v := &traverses.TraverseRenameTable{}
	(*tistmt).Accept(v)
	logics.LogicRenameTable(v, r.RuleHint)
}
