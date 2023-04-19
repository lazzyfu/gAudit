/*
@Time    :   2023/04/19 15:10:11
@Author  :   zongfei.fu 
@Desc    :   
*/


package inspect

import (
	"github.com/pingcap/tidb/parser/ast"
)

func AnalyzeTableRules() []Rule {
	return []Rule{
		{
			Hint:      "AnalyzeTable#检查",
			CheckFunc: (*Rule).RuleAnalyzeTable,
		},
	}
}

// RuleAnalyzeTable
func (r *Rule) RuleAnalyzeTable(tistmt *ast.StmtNode) {
	v := &TraverseAnalyzeTable{}
	(*tistmt).Accept(v)
	LogicAnalyzeTable(v, r)
}
