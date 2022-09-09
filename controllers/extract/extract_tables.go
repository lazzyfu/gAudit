/*
@Time    :   2022/09/09 10:35:02
@Author  :   zongfei.fu
@Desc    :   提取表名
*/

package extract

import (
	_ "embed"
	"fmt"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/controllers/parser"
	"sqlSyntaxAudit/forms"
	logger "sqlSyntaxAudit/middleware/log"

	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/sirupsen/logrus"
)

// 移除重复的值
func removeDuplicateElement(data []string) []string {
	result := make([]string, 0, len(data))
	temp := map[string]struct{}{}
	for _, item := range data {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// 返回数据
type ReturnData struct {
	Tables []string `json:"tables"` // 表名
	Type   string   `json:"type"`   // 语句类型
	Query  string   `json:"query"`  // 原始SQL
}

// 检查结构体
type Checker struct {
	Form  forms.ExtractTablesForm
	Audit *config.Audit
}

func (c *Checker) Extract(RequestID string) (error, []ReturnData) {
	var returnData []ReturnData
	err := c.Parse()
	if err != nil {
		logger.AppLog.WithFields(logrus.Fields{"request_id": RequestID}).Error(err)
		return err, returnData
	}
	for _, stmt := range c.Audit.TiStmt {
		var data ReturnData = ReturnData{Query: stmt.Text()}
		data.Tables, data.Type = ExtractTablesFromStatement(&stmt)
		returnData = append(returnData, data)
	}
	return nil, returnData
}

// 解析SQL语句
func (c *Checker) Parse() error {
	// 解析SQL
	var warns []error
	var err error
	// 解析
	c.Audit, warns, err = parser.NewParse(c.Form.SqlText, "", "")
	if len(warns) > 0 {
		return fmt.Errorf("Parse Warning: %s", utils.ErrsJoin("; ", warns))
	}
	if err != nil {
		return fmt.Errorf("sql解析错误:%s", err.Error())
	}
	return nil
}

// 提取表结构体
type ExtractTables struct {
	Tables []string // 表名
}

// 迭代select语句
func (e *ExtractTables) checkSelectItem(node ast.ResultSetNode) {
	if node == nil {
		return
	}
	// fmt.Println("类型: ", reflect.TypeOf(node))
	switch n := node.(type) {
	case *ast.SelectStmt:
		e.checkSubSelectItem(n)
	case *ast.Join:
		e.checkSelectItem(n.Left)
		e.checkSelectItem(n.Right)
	case *ast.TableSource:
		e.checkSelectItem(n.Source)
	case *ast.TableName:
		e.Tables = append(e.Tables, n.Name.String())
	}
}

// 迭代子查询
func (e *ExtractTables) checkSubSelectItem(node *ast.SelectStmt) {
	if node.From != nil {
		// 迭代from子查询
		e.checkSelectItem(node.From.TableRefs)
	}
	if node.Where != nil {
		e.checkExprItem(node.Where)
	}
	for _, item := range node.Fields.Fields {
		if item.Expr != nil {
			e.checkExprItem(item.Expr)
		}
	}
}

// 迭代表达式
func (e *ExtractTables) checkExprItem(expr ast.ExprNode) {
	switch ex := expr.(type) {
	case *ast.PatternInExpr:
		e.checkExprItem(ex.Sel)
	case *ast.CompareSubqueryExpr:
		e.checkExprItem(ex.R)
	case *ast.BinaryOperationExpr:
		e.checkExprItem(ex.L)
		e.checkExprItem(ex.R)
	case *ast.ExistsSubqueryExpr:
		e.checkExprItem(ex.Sel)
	case *ast.SubqueryExpr:
		e.checkSelectItem(ex.Query)
	}
}

// TraverseStatement
type TraverseStatement struct {
	Tables []string // 表名
	Type   string   // 语句类型
}

func (c *TraverseStatement) Enter(in ast.Node) (ast.Node, bool) {
	var e ExtractTables
	switch stmt := in.(type) {
	case *ast.SelectStmt:
		c.Type = "SELECT"
		e.checkSelectItem(stmt.From.TableRefs)
		e.checkExprItem(stmt.Where)
		if stmt.Having != nil {
			e.checkExprItem(stmt.Having.Expr)
		}
		for _, field := range stmt.Fields.Fields {
			e.checkExprItem(field.Expr)
		}
		if stmt.GroupBy != nil {
			for _, gb := range stmt.GroupBy.Items {
				e.checkExprItem(gb.Expr)
			}
		}
		c.Tables = append(c.Tables, e.Tables...)
	case *ast.InsertStmt:
		c.Type = "INSERT"
		if stmt.IsReplace {
			c.Type = "REPLACE"
		}
		e.checkSelectItem(stmt.Table.TableRefs)
		e.checkSelectItem(stmt.Select)
		c.Tables = append(c.Tables, e.Tables...)
	case *ast.UpdateStmt:
		c.Type = "UPDATE"
		e.checkSelectItem(stmt.TableRefs.TableRefs)
		c.Tables = append(c.Tables, e.Tables...)
	case *ast.DeleteStmt:
		c.Type = "DELETE"
		e.checkSelectItem(stmt.TableRefs.TableRefs)
		c.Tables = append(c.Tables, e.Tables...)
	case *ast.CreateTableStmt:
		c.Type = "CREATE TABLE"
		c.Tables = append(c.Tables, stmt.Table.Name.L)
	case *ast.CreateViewStmt:
		c.Type = "CREATE VIEW"
		c.Tables = append(c.Tables, stmt.ViewName.Name.L)
	case *ast.CreateIndexStmt:
		c.Type = "CREATE INDEX"
		c.Tables = append(c.Tables, stmt.Table.Name.L)
	case *ast.AlterTableStmt:
		c.Type = "ALTER TABLE"
		c.Tables = append(c.Tables, stmt.Table.Name.L)
	case *ast.DropIndexStmt:
		c.Type = "DROP INDEX"
		c.Tables = append(c.Tables, stmt.Table.Name.L)
	case *ast.RenameTableStmt:
		c.Type = "RENAME TABLE"
		for _, t := range stmt.TableToTables {
			c.Tables = append(c.Tables, t.OldTable.Name.String())
			c.Tables = append(c.Tables, t.NewTable.Name.String())
		}
	case *ast.DropTableStmt:
		c.Type = "DROP TABLE"
		for _, t := range stmt.Tables {
			c.Tables = append(c.Tables, t.Name.L)
		}
	case *ast.TruncateTableStmt:
		c.Type = "TRUNCATE TABLE"
		c.Tables = append(c.Tables, stmt.Table.Name.L)
	}
	return in, false
}

func (c *TraverseStatement) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func ExtractTablesFromStatement(stmt *ast.StmtNode) ([]string, string) {
	v := &TraverseStatement{}
	(*stmt).Accept(v)
	return removeDuplicateElement(v.Tables), v.Type
}
