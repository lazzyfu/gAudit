/*
@Time    :   2022/07/06 10:09:48
@Author  :   zongfei.fu
@Desc    :   路由
*/

package routers

import (
	views "sqlSyntaxAudit/views"

	"github.com/gin-gonic/gin"
)

func ApiRouterInit(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/audit", views.SyntaxInspect)
	}
}
