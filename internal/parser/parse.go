package parser

import (
	"github.com/lazzyfu/gaudit/internal/config"

	"github.com/pingcap/tidb/pkg/parser"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
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
