/*
@Time    :   2022/07/06 10:10:41
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sqlSyntaxAudit/common/kv"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/controllers/process"
	"sqlSyntaxAudit/forms"
	"sqlSyntaxAudit/global"
	logger "sqlSyntaxAudit/middleware/log"
	"strings"
	"time"

	"github.com/percona/go-mysql/query"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/sirupsen/logrus"
)

type Checker struct {
	Form      forms.SyntaxAudit
	Charset   string
	Collation string
	Audit     *config.Audit
	DB        *utils.DB
}

type ReturnData struct {
	Summary      []string `json:"summary"` // 规则摘要
	Level        string   `json:"level"`   // 提醒级别,INFO/WARN/ERROR
	AffectedRows int64    `json:"affected_rows"`
	Type         string   `json:"type"`
	FingerId     string   `json:"finger_id"`
	Query        string   `json:"query"` // 原始SQL
}

func (c *Checker) InitDB() {
	c.DB = &utils.DB{
		User:     c.Form.DbUser,
		Password: c.Form.DbPassword,
		Host:     c.Form.DbHost,
		Port:     c.Form.DbPort,
		Database: c.Form.DB,
		Timeout:  time.Duration(c.Form.Timeout),
	}
}

func (c *Checker) CustomParams() error {
	// 动态传参，当前请求传参可覆盖默认配置，仅对当前请求生效
	if len(c.Form.CustomAuditParams) == 0 {
		return nil
	}
	// 不允许接口传递的自定义的参数
	unAllowedCustomAuditParams := []string{
		"ListenAddress",
		"LogFilePath",
		"LogLevel",
	}

	// 验证传递的key是否在内置的参数内
	for key := range c.Form.CustomAuditParams {
		if utils.IsContain(unAllowedCustomAuditParams, key) {
			return fmt.Errorf("`custom_audit_parameters`不允许传递参数`%s`", key)
		}
		upperKey := strings.ToUpper(key) // 转换为大写
		rto := reflect.TypeOf(*global.App.AuditConfig)
		if _, ok := rto.FieldByName(upperKey); !ok {
			return fmt.Errorf("`custom_audit_parameters`传递的参数`%s`不存在", key)
		}
	}

	// map序列化
	data, _ := json.Marshal(c.Form.CustomAuditParams)
	r := bytes.NewReader([]byte(data))
	decoder := json.NewDecoder(r)

	// 动态参数赋值给默认模板
	// 优先级: post custom_audit_parameters > 自定义参数 > 内置默认参数
	if err := decoder.Decode(global.App.AuditConfig); err != nil {
		return err
	}
	return nil
}

func (c *Checker) CleanupAuditParams() {
	if global.App.AuditConfig.MAX_TABLE_NAME_LENGTH > 64 {
		global.App.AuditConfig.MAX_TABLE_NAME_LENGTH = 64
	}
	if global.App.AuditConfig.TABLE_COMMENT_LENGTH > 512 {
		global.App.AuditConfig.TABLE_COMMENT_LENGTH = 512
	}
	if global.App.AuditConfig.MAX_COLUMN_NAME_LENGTH > 64 {
		global.App.AuditConfig.MAX_COLUMN_NAME_LENGTH = 64
	}
	if global.App.AuditConfig.MAX_VARCHAR_LENGTH > 65535 {
		global.App.AuditConfig.MAX_COLUMN_NAME_LENGTH = 65535
	}
}

func (c *Checker) Parse() error {
	// 解析SQL
	var warns []error
	var err error
	// 处理自定义传参
	if err := c.CustomParams(); err != nil {
		return err
	}
	// 审核参数清洗
	c.CleanupAuditParams()
	// 解析
	c.Audit, warns, err = NewParse(c.Form.SqlText, c.Charset, c.Collation)
	if len(warns) > 0 {
		return fmt.Errorf("Parse Warning: %s", utils.ErrsJoin("; ", warns))
	}
	if err != nil {
		return fmt.Errorf("sql解析错误:%s", err.Error())
	}
	return nil
}

func (c *Checker) CreateTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// 建表语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	for _, rule := range CreateTableRules() {
		rule.DB = c.DB
		rule.KV = kv
		rule.CheckFunc(&rule, &stmt)
		if len(rule.Summary) > 0 {
			// 检查不通过
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.Summary...)
		}
		if rule.IsSkipNextStep {
			// 如果IsSkipNextStep为true，跳过接下来的检查步骤
			break
		}
	}
	return data
}

func (c *Checker) AlterTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) (ReturnData, string) {
	// alter语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	var mergeAlter string
	for _, rule := range AlterTableRules() {
		rule.DB = c.DB
		rule.KV = kv
		rule.CheckFunc(&rule, &stmt)
		if len(rule.MergeAlter) > 0 && len(mergeAlter) == 0 {
			mergeAlter = rule.MergeAlter
		}
		if len(rule.Summary) > 0 {
			// 检查不通过
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.Summary...)
		}
		if rule.IsSkipNextStep {
			fmt.Println("ggggg...........")
			// 如果IsSkipNextStep为true，跳过接下来的检查步骤
			break
		}
	}
	return data, mergeAlter
}

func (c *Checker) DropTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// drop/truncate语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	for _, rule := range DropTableRules() {
		rule.DB = c.DB
		rule.KV = kv
		rule.CheckFunc(&rule, &stmt)
		if len(rule.Summary) > 0 {
			// 检查不通过
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.Summary...)
		}
		if rule.IsSkipNextStep {
			// 如果IsSkipNextStep为true，跳过接下来的检查步骤
			break
		}
	}
	return data
}

func (c *Checker) DMLStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// delete/update/insert语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DML", Level: "INFO"}
	/*
		DML语句真的需要对同一个指纹的SQL跳过校验？
		1. DML规则并不多，对实际校验性能影响不大
		2. 每条DML都需要进行Explain，由于考虑传值不一样，因此指纹一样并不能代表Explain的影响行数一样
		3. 实际测试1000条update校验仅需800ms,2000条update校验仅需1500ms
		finger := kv.Get(fingerId)
		var IsSkipAudit bool
		if finger != nil {
			IsSkipAudit = true
		}
	*/
	for _, rule := range DMLRules() {
		rule.DB = c.DB
		rule.KV = kv
		rule.Query = stmt.Text()
		rule.CheckFunc(&rule, &stmt)
		data.AffectedRows = rule.AffectedRows
		if len(rule.Summary) > 0 {
			// 检查不通过
			data.Level = "WARN"
			data.Summary = append(data.Summary, rule.Summary...)
		}
		if rule.IsSkipNextStep {
			// 如果IsSkipNextStep为true，跳过接下来的检查步骤
			break
		}
	}
	return data
}

func (c *Checker) MergeAlter(kv *kv.KVCache, mergeAlters []string) ReturnData {
	// 检查merge操作
	var data ReturnData = ReturnData{Level: "INFO"}
	dbVersionIns := process.DbVersion{Version: kv.Get("dbVersion").(string)}
	if global.App.AuditConfig.ENABLE_MYSQL_MERGE_ALTER_TABLE && !dbVersionIns.IsTiDB() {
		if ok, _ := utils.IsRepeat(mergeAlters); ok {
			data.Summary = append(data.Summary, "MySQL同一张表的多个ALTER操作请合并为一条ALTER语句")
		}
	}
	if !global.App.AuditConfig.ENABLE_TIDB_MERGE_ALTER_TABLE && dbVersionIns.IsTiDB() {
		if ok, _ := utils.IsRepeat(mergeAlters); ok {
			data.Summary = append(data.Summary, "TiDB同一张表的多次ALTER操作请拆分为多条ALTER语句")
		}
	}
	if len(data.Summary) > 0 {
		data.Level = "WARN"
	}
	return data
}

func (c *Checker) Check(RequestID string) (err error, returnData []ReturnData) {
	c.InitDB()
	var mergeAlters []string // 存放alter语句中的表名
	// 解析SQL
	err = c.Parse()
	if err != nil {
		logger.AppLog.WithFields(logrus.Fields{"request_id": RequestID}).Error(err)
		return err, returnData
	}

	// 每次请求基于RequestID初始化kv cache
	kv := kv.NewKVCache(RequestID)
	// 获取目标数据库变量
	dbVars, err := GetDBVars(c.DB)
	if err != nil {
		errMsg := fmt.Sprintf("获取DB变量失败:%s", err.Error())
		logger.AppLog.WithFields(logrus.Fields{"request_id": RequestID}).Error(errMsg)
		return fmt.Errorf(errMsg), returnData
	}
	kv.Put("dbVersion", dbVars["dbVersion"])
	kv.Put("dbCharset", dbVars["dbCharset"])
	kv.Put("largePrefix", dbVars["largePrefix"])
	// 迭代stmt
	for _, stmt := range c.Audit.TiStmt {
		fingerId := query.Id(query.Fingerprint(stmt.Text()))
		kv.Put(fingerId, true)
		// 迭代
		switch stmt.(type) {
		case *ast.CreateTableStmt, *ast.CreateViewStmt:
			returnData = append(returnData, c.CreateTableStmt(stmt, kv, fingerId))
		case *ast.AlterTableStmt:
			data, mergeAlter := c.AlterTableStmt(stmt, kv, fingerId)
			mergeAlters = append(mergeAlters, mergeAlter)
			returnData = append(returnData, data)
		case *ast.DropTableStmt, *ast.TruncateTableStmt:
			returnData = append(returnData, c.DropTableStmt(stmt, kv, fingerId))
		case *ast.DeleteStmt, *ast.InsertStmt, *ast.UpdateStmt:
			returnData = append(returnData, c.DMLStmt(stmt, kv, fingerId))
		}
	}
	if len(mergeAlters) > 1 {
		mergeData := c.MergeAlter(kv, mergeAlters)
		if len(mergeData.Summary) > 0 {
			returnData = append(returnData, mergeData)
		}
	}
	return nil, returnData
}
