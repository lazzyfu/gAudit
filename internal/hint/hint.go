package hint

import (
	"github.com/lazzyfu/gaudit/internal/config"
	"github.com/lazzyfu/gaudit/internal/dao"
	"github.com/lazzyfu/gaudit/pkg/kv"
)

type RuleHint struct {
	Summary        []string `json:"summary"`       // 摘要
	AffectedRows   int      `json:"affected_rows"` // 默认为0
	IsSkipNextStep bool     // 是否跳过接下来的检查步骤
	DB             *dao.DB
	KV             *kv.KVCache[string]
	Query          string // 原始SQL
	MergeAlter     string
	AuditConfig    *config.AuditConfiguration
}
