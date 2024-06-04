/*
@Time    :   2022/07/06 10:10:14
@Author  :   xff
@Desc    :   None
*/

package views

import (
	"gAudit/controllers/checker"
	"gAudit/controllers/extract"
	"gAudit/forms"
	"gAudit/pkg/response"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// 语法检查
func SyntaxInspect(c *gin.Context) {
	var form *forms.SyntaxAuditForm

	if err := c.ShouldBind(&form); err != nil {
		response.ValidateFail(c, err.Error())
	} else {
		ch := checker.Checker{Form: form, RequestID: requestid.Get(c)}
		err, returnData := ch.Check()
		if err != nil {
			response.Fail(c, err.Error())
		} else {
			response.Success(c, returnData, "success")
		}
	}
}

// 提取表名
func ExtractTables(c *gin.Context) {
	var form *forms.ExtractTablesForm

	if err := c.ShouldBind(&form); err != nil {
		response.ValidateFail(c, err.Error())
	} else {
		checker := extract.Checker{Form: form, RequestID: requestid.Get(c)}
		err, returnData := checker.Extract()
		if err != nil {
			response.Fail(c, err.Error())
		} else {
			response.Success(c, returnData, "success")
		}
	}
}
