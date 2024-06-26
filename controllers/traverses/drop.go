/*
@Time    :   2022/06/24 13:12:20
@Author  :   xff
@Desc    :   遍历语法树,语法参考pingcap文档：https://github.com/pingcap/parser/blob/master/docs/quickstart.md
*/

package traverses

import (
	"github.com/pingcap/tidb/pkg/parser/ast"
)

// TraverseDropTable
type TraverseDropTable struct {
	Tables  []string // 表名
	IsMatch bool     // 是否匹配当前规则
}

func (c *TraverseDropTable) Enter(in ast.Node) (ast.Node, bool) {
	if stmt, ok := in.(*ast.DropTableStmt); ok {
		c.IsMatch = true
		for _, table := range stmt.Tables {
			c.Tables = append(c.Tables, table.Name.O)
		}
	}
	return in, false
}

func (c *TraverseDropTable) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

// TraverseTruncateTable
type TraverseTruncateTable struct {
	Table   string // 表名
	IsMatch bool   // 是否匹配当前规则
}

func (c *TraverseTruncateTable) Enter(in ast.Node) (ast.Node, bool) {
	if stmt, ok := in.(*ast.TruncateTableStmt); ok {
		c.IsMatch = true
		c.Table = stmt.Table.Name.O
	}
	return in, false
}

func (c *TraverseTruncateTable) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
