/*
@Time    :   2022/07/06 10:08:10
@Author  :   xff
@Desc    :   None
*/

package main

import (
	"flag"
	"fmt"
	"gAudit/bootstrap"
	"gAudit/global"
	"gAudit/middleware"
	"gAudit/routers"

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

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestid.New())
	r.Use(middleware.LoggerRequestToFile())

	// 路由
	routers.SetupRouter(r)

	// 启动
	err := r.Run(global.App.AuditConfig.ListenAddress)
	if err != nil {
		fmt.Println(err.Error())
	}
}
