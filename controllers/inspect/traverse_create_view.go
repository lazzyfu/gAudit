/*
@Time    :   2022/06/24 13:12:20
@Author  :   zongfei.fu
@Desc    :   遍历语法树,语法参考pingcap文档：https://github.com/pingcap/parser/blob/master/docs/quickstart.md
*/

package inspect

import (
	"github.com/pingcap/tidb/parser/ast"
)

// TraverseCreateViewIsExist
type TraverseCreateViewIsExist struct {
	View string // 视图
}

func (c *TraverseCreateViewIsExist) Enter(in ast.Node) (ast.Node, bool) {
	if stmt, ok := in.(*ast.CreateViewStmt); ok {
		c.View = stmt.ViewName.Name.String()
	}
	return in, false
}

func (c *TraverseCreateViewIsExist) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
