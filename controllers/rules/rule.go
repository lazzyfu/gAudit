/*
@Time    :   2022/06/29 15:30:31
@Author  :   zongfei.fu
*/

package rules

import (
	"gAudit/controllers"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

type Rule struct {
	*controllers.RuleHint
	Hint      string                     `json:"hint"` // 规则说明
	CheckFunc func(*Rule, *ast.StmtNode) // 函数名
}
