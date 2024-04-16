package controllers

import (
	"gAudit/config"
	"gAudit/controllers/dao"
	"gAudit/pkg/kv"
)

type RuleHint struct {
	Summary        []string `json:"summary"` // 摘要
	AffectedRows   int      `json:"affected_rows"`
	IsSkipNextStep bool     // 是否跳过接下来的检查步骤
	DB             *dao.DB
	KV             *kv.KVCache
	Query          string // 原始SQL
	MergeAlter     string
	AuditConfig    *config.AuditConfiguration
}
