/*
@Time    :   2022/08/25 16:41:19
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"github.com/pingcap/tidb/parser/ast"
)

type RenameTable struct {
	OldTable string // 表名
	NewTable string // 是否匹配当前规则
}

// TraverseRenameTable
type TraverseRenameTable struct {
	IsMatch int
	tables  []RenameTable
}

func (c *TraverseRenameTable) Enter(in ast.Node) (ast.Node, bool) {
	if stmt, ok := in.(*ast.RenameTableStmt); ok {
		c.IsMatch++
		for _, t := range stmt.TableToTables {
			c.tables = append(c.tables, RenameTable{
				OldTable: t.OldTable.Name.String(),
				NewTable: t.NewTable.Name.String(),
			})
		}
	}
	return in, false
}

func (c *TraverseRenameTable) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
