package global

import (
	"sqlSyntaxAudit/config"

	"gorm.io/gorm"
)

type Application struct {
	DB          *gorm.DB
	AuditConfig *config.AuditConfiguration
}

var App = new(Application)
