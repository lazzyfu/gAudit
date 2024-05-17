/*
@Time    :   2022/07/06 10:09:48
@Author  :   xff
@Desc    :   路由
*/

package routers

import (
	"gAudit/views"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) *gin.Engine {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/audit", views.SyntaxInspect)
		v1.POST("/extract-tables", views.ExtractTables)
	}
	return r
}
