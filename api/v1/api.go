package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/lazzyfu/gaudit/api/v1/handlers"
)

func SetupRouter(r *gin.Engine) *gin.Engine {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/audit", handlers.SyntaxInspect)
		v1.POST("/extract-tables", handlers.ExtractTables)
	}
	return r
}
