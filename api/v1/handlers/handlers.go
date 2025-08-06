package handlers

import (
	"github.com/lazzyfu/gaudit/internal/checker"
	"github.com/lazzyfu/gaudit/internal/extract"

	"github.com/lazzyfu/gaudit/forms"
	"github.com/lazzyfu/gaudit/pkg/response"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// 语法检查
func SyntaxInspect(c *gin.Context) {
	var form forms.SyntaxAuditForm

	if err := c.ShouldBindJSON(&form); err != nil {
		response.ValidateFail(c, err.Error())
		return
	}

	ch := checker.Checker{Form: &form, RequestID: requestid.Get(c)}
	err, returnData := ch.Check()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, returnData, "success")
}

func ExtractTables(c *gin.Context) {
	var form forms.ExtractTablesForm

	if err := c.ShouldBindJSON(&form); err != nil {
		response.ValidateFail(c, err.Error())
		return
	}

	checker := extract.Checker{Form: &form, RequestID: requestid.Get(c)}
	err, returnData := checker.Extract()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, returnData, "success")
}
