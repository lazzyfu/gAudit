/*
@Time    :   2022/08/16 16:10:45
@Author  :   zongfei.fu
@Desc    :   None

可以自定义测试语句
# 运行测试
 ~/Desktop/github/gAudit/tests/ [main*] go test -v audit_test.go
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] POST   /api/v1/audit             --> sqlSyntaxAudit/views.SyntaxInspect (4 handlers)
=== RUN   TestPost
# AlterTable#检测
  *  表`t1`不存在

# CreateTable#检查表是否存在
  *  表`c1`的自增列初始值必须显式指定且设置为1【例如:AUTO_INCREMENT=1】
  *  表`c1`的主键id必须使用bigint类型
  *  表`c1`未定义字段类型为DEFAULT CURRENT_TIMESTAMP的审计字段【例如:CREATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'】
  *  表`c1`未定义字段类型为DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP的审计字段【例如:UPDATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'】

# CreateTable#检查CreateTableAs语法
  *  不允许使用create table as语法[表`t1`]

# CreateTable#检查CreateTableLike语法
  *  不允许使用create table like语法[表`t1`]

# CreateView#检查是否允许创建视图

# CreateTable#表检测
  *  表`test1`指定的存储引擎`MyISAM`不符合要求,支持的存储引擎为`[InnoDB]`
  *  表`test1`必须要有注释
  *  表`test1`指定的字符集排序规则`utf8`不符合要求,应指定前缀为utf8_的排序规则,推荐的字符集排序规则为utf8_general_ci
  *  表`test1`未定义字段类型为DEFAULT CURRENT_TIMESTAMP的审计字段【例如:CREATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'】
  *  列`env`不能定义`NOT NULL DEFAULT NULL`[表`test1`]
  *  列`addr`必须要有注释[表`test1`]
  *  列`addr`最大允许定义的varchar长度为8096,当前varchar长度为30000[表`test1`]
  *  列`addr`必须定义为`NOT NULL`[表`test1`]
  *  列`finish_at`需要设置一个默认值[表`test1`]
  *  列`addr`必须同时指定字符集和排序规则[表`test1`]
  *  二级索引前缀不符合要求,必须以`IDX_`开头(不区分大小写)[表`test1`]
  *  表`test1`的索引`idx_addr`超出了innodb-large-prefix限制,当前索引长度为90002字节,最大限制为3072字节【例如:可使用前缀索引,如:Field(length)】

--- PASS: TestPost (0.04s)
PASS
ok      command-line-arguments  0.687s
*/

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/global"
	logger "sqlSyntaxAudit/middleware/log"
	"sqlSyntaxAudit/routers"
	"testing"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

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
	Summary      []string `json:"summary"` // 规则摘要
	Level        string   `json:"level"`   // 提醒级别,INFO/WARN/ERROR
	AffectedRows int      `json:"affected_rows"`
	Type         string   `json:"type"`
	FingerId     string   `json:"finger_id"`
	Query        string   `json:"query"` // 原始SQL
}

type resultResponse struct {
	RequestID string         `json:"request_id"`
	Code      string         `json:"code"`
	Data      []dataResponse `json:"data"`
	Message   string         `json:"message"`
}

func SqlTexts() map[string]string {
	RuleCreateTableIsExist := `CREATE TABLE c1 (
		id tinyint(3) unsigned NOT NULL AUTO_INCREMENT,
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='测试表'`

	RuleCreateTableAs := `create table t1 as select * from c1;`
	RuleCreateTableLike := `create table t1 like c1;`
	RuleCreateViewIsExist := `create view v1 as select id from t1`
	RuleCreateTable := `CREATE TABLE test1 (
		id bigint unsigned NOT NULL AUTO_INCREMENT,
		env varchar(32) NOT NULL DEFAULT null comment '环境' ,
		addr varchar(30000)   collate utf8_unicode_ci default null,
		finish_at datetime null comment '完成时间',
		D_UPDATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
		PRIMARY KEY (id),
		key idx_addr (addr),
		key unxx_env (env, addr(10))
	) ENGINE=MyISAM AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COLLATE=utf8mb4_unicode_ci `
	RuleAlterTable := `alter table t1 modify name varchar(255) not null default ''`

	return map[string]string{
		"CreateTable#检查表是否存在":             RuleCreateTableIsExist,
		"CreateTable#检查CreateTableAs语法":   RuleCreateTableAs,
		"CreateTable#检查CreateTableLike语法": RuleCreateTableLike,
		"CreateView#检查是否允许创建视图":           RuleCreateViewIsExist,
		"CreateTable#表检测":                 RuleCreateTable,
		"AlterTable#检测":                   RuleAlterTable,
	}
}

func TestPost(t *testing.T) {
	for key, value := range SqlTexts() {
		form := map[string]interface{}{
			"db_user":     "sqlsyntaxaudit_rw",
			"db_password": "1234.com",
			"db_host":     "127.0.0.1",
			"db_port":     3306,
			"db":          "test",
			"timeout":     3000,
			"sqltext":     value,
		}
		w := httptest.NewRecorder()
		jdata, err := json.Marshal(form)
		if err != nil {
			t.Error(err)
		}

		req, _ := http.NewRequest("POST", "http://127.0.0.1:8082/api/v1/audit", bytes.NewReader(jdata))
		req.Header.Set("Content-Type", "application/json;charset=utf-8")

		router.ServeHTTP(w, req)
		var result *resultResponse
		err = json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Error(err)
		}
		var summary []string
		for _, row := range result.Data {
			summary = append(summary, row.Summary...)
		}
		fmt.Println("#", key)
		for _, s := range summary {
			fmt.Println("  * ", s)
		}

		fmt.Println("")
	}
}
