/*
@Time    :   2022/06/23 14:24:59
@Author  :   zongfei.fu
@Desc    :   封装响应
*/

package response

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type Response struct {
	RequestID string      `json:"request_id"`
	Code      string      `json:"code"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		requestid.Get(c),
		"0000",
		data,
		"success",
	})
}

func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		requestid.Get(c),
		"0001",
		nil,
		msg,
	})
}

func ValidateFail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		requestid.Get(c),
		"0002",
		nil,
		msg,
	})
}
