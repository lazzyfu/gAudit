/*
@Time    :   2022/07/06 10:12:42
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"sqlSyntaxAudit/controllers/extract"

	"github.com/pingcap/tidb/parser/ast"
)

func CreateViewRules() []Rule {
	return []Rule{
		{
			Hint:      "CreateView#检查视图是否存在",
			CheckFunc: (*Rule).RuleCreateViewIsExist,
		},
	}
}

// RuleCreateViewIsExist
func (r *Rule) RuleCreateViewIsExist(tistmt *ast.StmtNode) {
	v := &TraverseCreateViewIsExist{}
	(*tistmt).Accept(v)
	v.Tables, _ = extract.ExtractTablesFromStatement(tistmt)
	LogicCreateViewIsExist(v, r)
}
