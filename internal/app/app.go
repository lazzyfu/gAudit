package app

import (
	"flag"
	"fmt"
	"time"

	"github.com/lazzyfu/gaudit/internal/config"
	"github.com/lazzyfu/gaudit/internal/global"
	"github.com/lazzyfu/gaudit/internal/log"
	"github.com/lazzyfu/gaudit/internal/server"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func Run() {
	// 解析输入参数
	configFile := flag.String("config", "./config.json", "审核参数配置文件")
	logPath := flag.String("logpath", "./", "日志文件路径")
	flag.Parse()

	// 初始化应用日志（优先初始化，保证后续日志输出）

	logFile := time.Now().Format("2006-01-02") + ".app.log"
	global.App.Log = log.InitLogger(*logPath, logFile)

	// 初始化配置
	auditConfig, err := config.InitializeAuditConfig(*configFile)
	if err != nil {
		global.App.Log.Errorf("启动服务失败: %v", err)
		panic(fmt.Sprintf("无法加载配置文件: %s, 错误: %v", *configFile, err))
	}
	global.App.AuditConfig = auditConfig
	// global.App.Log.Infof("加载配置文件: +%v 成功", *global.App.AuditConfig)

	// gin框架
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestid.New())

	// 初始化请求日志
	r.Use(log.LoggerRequestToFile(global.App.Log))

	// 路由
	server.SetupRouter(r)

	// 启动
	if err := r.Run(global.App.AuditConfig.ListenAddress); err != nil {
		global.App.Log.Errorf("启动服务失败: %v", err)
	}
}
