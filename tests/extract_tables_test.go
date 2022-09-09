/*
@Time    :   2022/08/16 16:10:45
@Author  :   zongfei.fu
@Desc    :   None

可以自定义测试语句
# 运行测试
 ~/Desktop/github/gAudit/tests/ [main*] go test -v extract_tables_test.go
*/

package tests

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/global"
	logger "sqlSyntaxAudit/middleware/log"
	"sqlSyntaxAudit/routers"
	"strings"
	"testing"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

//go:embed extract_tables_template.sql
var sqltext string

var router *gin.Engine

func init() {
	// 初始化配置
	global.App.AuditConfig = config.InitializeAuditConfig("../template/config.json")

	r := gin.New()

	r.Use(gin.Recovery())
	logger.Setup()
	r.Use(requestid.New())
	r.Use(logger.LoggerToFile())

	// 路由
	router = routers.SetupRouter(r)
}

type dataResponse struct {
	Tables []string `json:"tables"` // 表名
	Type   string   `json:"type"`
	Query  string   `json:"query"` // 原始SQL
}

type resultResponse struct {
	RequestID string         `json:"request_id"`
	Code      string         `json:"code"`
	Data      []dataResponse `json:"data"`
	Message   string         `json:"message"`
}

func TestExtractTablesPost(t *testing.T) {
	form := map[string]interface{}{
		"sqltext": sqltext,
	}
	w := httptest.NewRecorder()
	jdata, err := json.Marshal(form)
	if err != nil {
		t.Error(err)
	}

	req, _ := http.NewRequest("POST", "http://127.0.0.1:8082/api/v1/extract-tables", bytes.NewReader(jdata))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	router.ServeHTTP(w, req)
	var result *resultResponse
	// fmt.Println(w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Error(err)
	}
	// fmt.Println(result.Data)
	for _, row := range result.Data {
		fmt.Println("表名: ", strings.Join(row.Tables, ", "))
	}
}
