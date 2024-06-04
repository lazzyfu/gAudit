/*
@Time    :   2022/07/06 10:08:10
@Author  :   xff
*/

package main

import (
	"flag"
	"fmt"
	"gAudit/bootstrap"
	"gAudit/global"
	"gAudit/middleware"
	"gAudit/routers"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// 接收输入
var configFile = flag.String("config", "./config.json", "审核参数配置文件")

func main() {
	// 解析输入的参数
	flag.Parse()

	// 初始化配置
	global.App.AuditConfig = bootstrap.InitializeAuditConfig(*configFile)

	// 初始化日志
	bootstrap.InitializeLog()

	// gin框架
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestid.New())

	// 初始化请求日志记录器
	requestLogger := middleware.InitLogger(time.Now().Format("2006-01-02") + "-request.log")
	r.Use(middleware.LoggerRequestToFile(requestLogger))

	// 路由
	routers.SetupRouter(r)

	// 启动
	err := r.Run(global.App.AuditConfig.ListenAddress)
	if err != nil {
		fmt.Println(err.Error())
	}
}
