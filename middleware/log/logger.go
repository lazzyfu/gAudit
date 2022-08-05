/*
@Time    :   2022/07/18 14:54:37
@Author  :   zongfei.fu
@Desc    :   None

使用方法：
import logger "sqlSyntaxAudit/middleware/log"
logger.AppLog.Error(errMsg)
*/

package logger

import (
	"fmt"
	"os"
	"path"
	"sqlSyntaxAudit/global"
	"strings"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var AppLog *logrus.Logger
var WebLog *logrus.Logger

func Setup() {
	initAppLog()
	initWebLog()
}

func initAppLog() {
	now := time.Now()
	logFileName := now.Format("2006-01-02") + ".app.log"
	AppLog = initLog(logFileName)
}

func initWebLog() {
	now := time.Now()
	logFileName := now.Format("2006-01-02") + ".web.log"
	WebLog = initLog(logFileName)
}

func initLog(logFileName string) *logrus.Logger {
	logFilePath := global.App.AuditConfig.LogFilePath
	if err := os.MkdirAll(logFilePath, 0777); err != nil {
		fmt.Println(err.Error())
	}
	// 日志文件
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			fmt.Println(err.Error())
		}
	}
	// 写入文件
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("err", err)
	}

	// 实例化
	logger := logrus.New()

	// 设置输出
	logger.Out = src

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

func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := fmt.Sprintf("%.3dms", endTime.Sub(startTime)/1e6)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		//日志格式
		var RequestID string = requestid.Get(c)
		WebLog.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
			"request_id":   RequestID,
		}).Info()
	}
}
