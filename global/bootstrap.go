package global

import (
	"log"
	"sqlSyntaxAudit/models"
)

func InitTables() {
	err := App.DB.AutoMigrate(&models.TestCase{})
	if err != nil {
		log.Fatal(err.Error())
	}
}
