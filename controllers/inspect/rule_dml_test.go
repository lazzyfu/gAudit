/*
@Time    :   2022/09/13 16:20:28
@Author  :   zongfei.fu
@Desc    :   None

# 测试
 ~/Desktop/github/gAudit/controllers/inspect/ [main*] go test -v
=== RUN   TestRuleDML
=== RUN   TestRuleDML/限制部分表进行语法审核
=== RUN   TestRuleDML/是否允许INSERT_INTO_SELECT语法
--- PASS: TestRuleDML (0.01s)
    --- PASS: TestRuleDML/限制部分表进行语法审核 (0.01s)
    --- PASS: TestRuleDML/是否允许INSERT_INTO_SELECT语法 (0.00s)
PASS
ok      sqlSyntaxAudit/controllers/inspect      1.040s
*/

package inspect

import (
	"crypto/rand"
	"fmt"
	"log"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/forms"
	"sqlSyntaxAudit/global"
	logger "sqlSyntaxAudit/middleware/log"
	"sqlSyntaxAudit/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
# 指定本地数据库账号，下面部分测试用例需要连接到本地数据库
# 创建本地测试账号和库
create user 'sqlsyntaxaudit_rw'@'%' identified by '1234.com';
create database test;
grant all on test.* to 'sqlsyntaxaudit_rw'@'%';
*/
var (
	DbUser     = "sqlsyntaxaudit_rw"
	DbPassword = "1234.com"
	DbHost     = "127.0.0.1"
	DbPort     = 3306
	DB         = "test"
)

func init() {
	// 初始化配置
	global.App.AuditConfig = &config.AuditConfiguration{
		LogFilePath: "../../logs",
	}
	logger.Setup()
	// 初始化配置
	global.App.AuditConfig = config.InitializeAuditConfig("../../template/config.json")
	// 初始化测试表
	var err error
	global.App.DB, err = models.InitDB(DbUser, DbPassword, DbHost, DbPort, DB)
	if err != nil {
		log.Fatal(err)
	}
	global.InitTables()
	// 插入测试数据
	global.App.DB.Exec("delete from test_case")
	global.App.DB.Model(&models.TestCase{}).Create([]map[string]interface{}{
		{"ID": 1, "Env": "prod", "ClusterName": "orc_tt1", "Datacenter": "hw", "Region": "z1", "Hostname": "test_host_1", "Port": 3306, "PromotionRule": "prefer"},
		{"ID": 2, "Env": "prod", "ClusterName": "orc_tt1", "Datacenter": "hw", "Region": "z2", "Hostname": "test_host_2", "Port": 3306, "PromotionRule": "neutral"},
		{"ID": 3, "Env": "prod", "ClusterName": "orc_tt1", "Datacenter": "hw", "Region": "z3", "Hostname": "test_host_3", "Port": 3306, "PromotionRule": "neutral"},
	})
}

func GetRandomString2(n int) string {
	randBytes := make([]byte, n/2)
	_, _ = rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}

func TestRuleDML(t *testing.T) {
	tests := []struct {
		name    string
		form    forms.SyntaxAuditForm
		wantErr error
		wantRes []ReturnData
	}{
		{
			name: "限制部分表进行语法审核",
			form: forms.SyntaxAuditForm{
				CustomAuditParams: map[string]interface{}{
					"DISABLE_AUDIT_DML_TABLES": []config.DisableTablesAudit{
						{
							DB:     "test",
							Tables: []string{"test_case"},
							Reason: "研发禁止审核和提交",
						},
					},
				},
				SqlText: "delete from test_case where id > 1",
			},
			wantRes: []ReturnData{
				{
					Summary:      []string{"表`test`.`test_case`被限制进行DML语法审核,原因: 研发禁止审核和提交"},
					Level:        "WARN",
					AffectedRows: 0,
					Type:         "DML",
					FingerId:     "D3A87C5D8BFAE066",
					Query:        "delete from test_case where id > 1",
				},
			},
		},
		{
			name: "检查表是否存在",
			form: forms.SyntaxAuditForm{
				CustomAuditParams: map[string]interface{}{},
				SqlText:           "delete from test_case1",
			},
			wantRes: []ReturnData{
				{
					Summary:      []string{"表或视图`test_case1`不存在"},
					Level:        "WARN",
					AffectedRows: 0,
					Type:         "DML",
					FingerId:     "3709CBCBC14B50C2",
					Query:        "delete from test_case1",
				},
			},
		},
		{
			name: "不允许INSERT INTO SELECT语法",
			form: forms.SyntaxAuditForm{
				CustomAuditParams: map[string]interface{}{"DISABLE_INSERT_INTO_SELECT": true},
				SqlText:           "insert into test_case select 1",
			},
			wantRes: []ReturnData{
				{
					Summary:      []string{"禁止使用INSERT into select语法"},
					Level:        "WARN",
					AffectedRows: 0,
					Type:         "DML",
					FingerId:     "A9CDEDF0B97E0AC2",
					Query:        "insert into test_case select 1",
				},
			},
		},
		{
			name: "不允许insert/replace into on duplicate语法语法",
			form: forms.SyntaxAuditForm{
				CustomAuditParams: map[string]interface{}{"DISABLE_ON_DUPLICATE": true},
				SqlText:           "insert test_case(`id`, `env`, `cluster_name`) values(3, 'test', 'orc_yy1') ON DUPLICATE KEY UPDATE cluster_name='orc_yy1'",
			},
			wantRes: []ReturnData{
				{
					Summary:      []string{"禁止使用INSERT into on duplicate语法"},
					Level:        "WARN",
					AffectedRows: 0,
					Type:         "DML",
					FingerId:     "CB42BF6919EE10DA",
					Query:        "insert test_case(`id`, `env`, `cluster_name`) values(3, 'test', 'orc_yy1') ON DUPLICATE KEY UPDATE cluster_name='orc_yy1'",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 本地测试账号密码
			tt.form.DbUser = DbUser
			tt.form.DbPassword = DbPassword
			tt.form.DbHost = DbHost
			tt.form.DbPort = DbPort
			tt.form.DB = DB

			checker := Checker{Form: tt.form}
			err, res := checker.Check(GetRandomString2(24))
			fmt.Println("实际输出:", res)
			fmt.Println("预期输出:", tt.wantRes)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				// 预期会有错误返回，就不需要进一步校验res了
				return
			}
			assert.Equal(t, tt.wantRes, res)
		})
	}
}
