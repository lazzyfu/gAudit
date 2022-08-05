package global

import (
	"log"
	"sqlSyntaxAudit/models"
)

func InitTables() {
	err := App.DB.AutoMigrate(&models.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}
}
