package rules

import (
	"github.com/lazzyfu/gaudit/internal/hint"
	"github.com/pingcap/tidb/pkg/parser/ast"
)

type Rule struct {
	*hint.RuleHint
	Hint      string                     `json:"hint"` // 规则说明
	CheckFunc func(*Rule, *ast.StmtNode) // 函数名
}
