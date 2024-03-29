/*
@Time    :   2022/07/06 10:10:14
@Author  :   zongfei.fu
@Desc    :   None
*/

package views

import (
	"sqlSyntaxAudit/common/response"
	"sqlSyntaxAudit/controllers/extract"
	"sqlSyntaxAudit/controllers/inspect"
	"sqlSyntaxAudit/forms"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func SyntaxInspect(c *gin.Context) {
	var form forms.SyntaxAuditForm
	var RequestID string = requestid.Get(c)
	form.RequestID = RequestID

	if err := c.ShouldBind(&form); err != nil {
		response.ValidateFail(c, err.Error())
	} else {
		checker := inspect.Checker{Form: form}
		err, returnData := checker.Check(RequestID)
		if err != nil {
			response.Fail(c, err.Error())
		} else {
			response.Success(c, returnData)
		}
	}
}

// 提取表名
func ExtractTables(c *gin.Context) {
	var form forms.ExtractTablesForm
	var RequestID string = requestid.Get(c)
	form.RequestID = RequestID

	if err := c.ShouldBind(&form); err != nil {
		response.ValidateFail(c, err.Error())
	} else {
		checker := extract.Checker{Form: form}
		err, returnData := checker.Extract(RequestID)
		if err != nil {
			response.Fail(c, err.Error())
		} else {
			response.Success(c, returnData)
		}
	}
}
