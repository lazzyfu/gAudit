/*
@Time    :   2022/07/18 14:54:37
@Author  :   xff
*/

package middleware

import (
	"fmt"
	"gAudit/global"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 初始化日志记录器，支持日志文件轮转
func InitLogger(logFileName string) *logrus.Logger {
	// 创建日志文件夹
	logFilePath := global.App.AuditConfig.LogFilePath
	if err := os.MkdirAll(logFilePath, 0777); err != nil {
		fmt.Println(err.Error())
	}

	// 实例化
	logger := logrus.New()

	// 使用lumberjack进行日志轮转
	logger.SetOutput(&lumberjack.Logger{
		Filename:   path.Join(logFilePath, logFileName),
		MaxSize:    100,  // 每个日志文件的最大尺寸（单位：MB）
		MaxBackups: 30,   // 保留的旧日志文件的最大数量
		MaxAge:     7,    // 保留旧日志文件的最大天数
		Compress:   true, // 是否压缩旧的日志文件
	})

	// 设置日志级别
	switch strings.ToLower(global.App.AuditConfig.LogLevel) {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// 设置日志格式
	logger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	}
	return logger
}

// LoggerRequestToFile 是Gin的中间件，用于记录请求日志到文件
func LoggerRequestToFile(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 继续处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latencyTime := fmt.Sprintf("%dms", endTime.Sub(startTime).Milliseconds())

		//日志格式
		logger.WithFields(logrus.Fields{
			"status_code":       c.Writer.Status(),
			"latency_time":      latencyTime,
			"request_client_ip": c.ClientIP(),
			"request_method":    c.Request.Method,
			"request_uri":       c.Request.RequestURI,
			"request_ua":        c.Request.UserAgent(),
			"request_referer":   c.Request.Referer(),
			"request_id":        requestid.Get(c),
			"request_header":    c.Request.Header,
		}).Info("HTTP request logged")
	}
}
