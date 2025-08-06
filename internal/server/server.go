package server

import (
	"github.com/gin-gonic/gin"
	apiV1 "github.com/lazzyfu/gaudit/api/v1"
)

func SetupRouter(r *gin.Engine) *gin.Engine {
	// 注册 v1 路由
	apiV1.SetupRouter(r)
	return r
}
