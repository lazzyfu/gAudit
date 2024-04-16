package global

import (
	"gAudit/config"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Application struct {
	DB          *gorm.DB
	AuditConfig *config.AuditConfiguration
	Log         *logrus.Logger
}

var App = new(Application)
