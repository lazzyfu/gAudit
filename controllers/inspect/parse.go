/*
@Time    :   2022/07/06 10:12:33
@Author  :   zongfei.fu 
@Desc    :   None
*/

package inspect

import (
	"sqlSyntaxAudit/config"

	"github.com/pingcap/parser"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

// NewParse
func NewParse(sqltext, charset, collation string) (*config.Audit, []error, error) {
	q := &config.Audit{Query: sqltext}

	// tidb parser 语法解析
	var warns []error
	var err error
	q.TiStmt, warns, err = parser.New().Parse(sqltext, charset, collation)
	return q, warns, err
}
