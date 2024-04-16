/*
@Time    :   2022/07/06 10:10:14
@Author  :   zongfei.fu
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

func SyntaxInspect(c *gin.Context) {
	var form forms.SyntaxAuditForm
	form.RequestID = requestid.Get(c)

	if err := c.ShouldBind(&form); err != nil {
		response.ValidateFail(c, err.Error())
	} else {
		ch := checker.Checker{Form: form}
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
			response.Success(c, returnData, "success")
		}
	}
}
