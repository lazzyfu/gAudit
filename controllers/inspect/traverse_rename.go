/*
@Time    :   2022/08/25 16:41:19
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"github.com/pingcap/parser/ast"
)

// TraverseRenameTable
type TraverseRenameTable struct {
	IsMatch  int
	OldTable string // 表名
	NewTable string // 是否匹配当前规则
}

func (c *TraverseRenameTable) Enter(in ast.Node) (ast.Node, bool) {
	if stmt, ok := in.(*ast.RenameTableStmt); ok {
		c.IsMatch++
		c.OldTable = stmt.OldTable.Name.String()
		c.NewTable = stmt.NewTable.Name.String()
	}
	return in, false
}

func (c *TraverseRenameTable) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
