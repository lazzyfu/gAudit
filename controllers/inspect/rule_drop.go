/*
@Time    :   2022/07/06 10:12:58
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"github.com/pingcap/tidb/parser/ast"
)

func DropTableRules() []Rule {
	return []Rule{
		{
			Hint:      "DropTable#检查",
			CheckFunc: (*Rule).RuleDropTable,
		},
		{
			Hint:      "TruncateTable#检查",
			CheckFunc: (*Rule).RuleTruncateTable,
		},
	}
}

// RuleDropTable
func (r *Rule) RuleDropTable(tistmt *ast.StmtNode) {
	v := &TraverseDropTable{}
	(*tistmt).Accept(v)
	LogicDropTable(v, r)
}

// RuleTruncateTable
func (r *Rule) RuleTruncateTable(tistmt *ast.StmtNode) {
	v := &TraverseTruncateTable{}
	(*tistmt).Accept(v)
	LogicTruncateTable(v, r)
}
