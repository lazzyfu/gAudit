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
)

func InitLog(logFileName string) *logrus.Logger {
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

func LoggerRequestToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		latencyTime := fmt.Sprintf("%dms", endTime.Sub(startTime).Milliseconds())

		//日志格式
		global.App.Log.WithFields(logrus.Fields{
			"status_code":       c.Writer.Status(),
			"latency_time":      latencyTime,
			"request_client_ip": c.ClientIP(),
			"request_method":    c.Request.Method,
			"request_uri":       c.Request.RequestURI,
			"request_ua":        c.Request.UserAgent(),
			"request_referer":   c.Request.Referer(),
			"request_id":        requestid.Get(c),
			"request_header":    c.Request.Header,
		}).Info()
	}
}
