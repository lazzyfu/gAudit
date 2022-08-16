/*
@Time    :   2022/07/06 10:08:10
@Author  :   zongfei.fu
@Desc    :   None
*/

package main

import (
	"flag"
	"fmt"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/global"
	logger "sqlSyntaxAudit/middleware/log"
	"sqlSyntaxAudit/routers"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// 接收输入
var configFile = flag.String("config", "./config.json", "审核参数配置文件")

func main() {
	// 解析输入的参数
	flag.Parse()

	// 初始化配置
	global.App.AuditConfig = config.InitializeAuditConfig(*configFile)

	r := gin.New()

	r.Use(gin.Recovery())
	logger.Setup()
	r.Use(requestid.New())
	r.Use(logger.LoggerToFile())

	// 路由
	routers.SetupRouter(r)

	// 启动
	err := r.Run(global.App.AuditConfig.ListenAddress)
	if err != nil {
		fmt.Println(err.Error())
	}
}
