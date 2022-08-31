/*
@Time    :   2022/06/29 15:30:31
@Author  :   zongfei.fu
*/

package inspect

import (
	"sqlSyntaxAudit/common/kv"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/config"

	"github.com/pingcap/tidb/parser/ast"
)

type Rule struct {
	Hint           string   `json:"hint"`    // 规则说明
	Summary        []string `json:"summary"` // 规则摘要
	AffectedRows   int      `json:"affected_rows"`
	IsSkipNextStep bool     // 是否跳过接下来的检查步骤
	DB             *utils.DB
	KV             *kv.KVCache
	Query          string // 原始SQL
	MergeAlter     string
	AuditConfig    *config.AuditConfiguration

	CheckFunc func(*Rule, *ast.StmtNode) // 函数名
}
