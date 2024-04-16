/*
@Author  :   lazzyfu
@Desc    :   response
*/

package response

import (
	"gAudit/global"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// code 字段是表示请求状态
// code 的值为0000，表示请求成功；code 的值为0001，表示请求失败
type Response struct {
	RequestID string      `json:"request_id"`
	Code      string      `json:"code"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
}

func writeResponse(c *gin.Context, code string, data interface{}, msg string) {
	requestID := requestid.Get(c)
	if code == "0000" {
		global.App.Log.WithFields(logrus.Fields{"request_id": requestID, "type": "response"}).Info(msg)
	} else {
		global.App.Log.WithFields(logrus.Fields{"request_id": requestID, "type": "response"}).Error(msg)
	}

	c.JSON(http.StatusOK, Response{requestID, code, data, msg})
}

func Success(c *gin.Context, data interface{}, msg string) {
	writeResponse(c, "0000", data, msg)
}

func Fail(c *gin.Context, msg string) {
	writeResponse(c, "0001", nil, msg)
}

func ValidateFail(c *gin.Context, msg string) {
	writeResponse(c, "0001", nil, msg)
}
