/*
@Time    :   2022/07/06 10:10:41
@Author  :   xff
@Desc    :   None
*/

package checker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gAudit/config"
	"gAudit/controllers"
	"gAudit/controllers/dao"
	"gAudit/controllers/parser"
	"gAudit/controllers/process"
	"gAudit/controllers/rules"
	"gAudit/forms"
	"gAudit/global"
	"gAudit/pkg/kv"
	"gAudit/pkg/utils"
	"reflect"
	"regexp"
	"strings"
	"time"

	query "gAudit/pkg/query"

	"github.com/jinzhu/copier"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
	"github.com/sirupsen/logrus"
)

// 返回数据
type ReturnData struct {
	Summary      []string `json:"summary"` // 规则摘要
	Level        string   `json:"level"`   // 提醒级别,INFO/WARN/ERROR
	AffectedRows int      `json:"affected_rows"`
	Type         string   `json:"type"`
	FingerId     string   `json:"finger_id"`
	Query        string   `json:"query"` // 原始SQL
}
type Checker struct {
	Form        forms.SyntaxAuditForm
	Charset     string
	Collation   string
	Audit       *config.Audit
	DB          *dao.DB
	AuditConfig config.AuditConfiguration
}

// 初始化DB
func (c *Checker) InitDB() {
	c.DB = &dao.DB{
		User:     c.Form.DbUser,
		Password: c.Form.DbPassword,
		Host:     c.Form.DbHost,
		Port:     c.Form.DbPort,
		Database: c.Form.DB,
		Timeout:  time.Duration(c.Form.Timeout),
	}
}

// 动态传参，当前请求传参可覆盖默认配置，仅对当前请求生效
func (c *Checker) CustomParams() error {
	// 赋值给新变量，使用copier进行深copy，会一层一层进行copy
	err := copier.CopyWithOption(&c.AuditConfig, global.App.AuditConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return fmt.Errorf("审核参数解析错误`%s`不存在", err)
	}

	// 判断传入是否为空
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
		rto := reflect.TypeOf(c.AuditConfig)
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
	if err := decoder.Decode(&c.AuditConfig); err != nil {
		return err
	}

	// 拦截异常传参
	func() {
		if c.AuditConfig.MAX_TABLE_NAME_LENGTH > 64 {
			c.AuditConfig.MAX_TABLE_NAME_LENGTH = 64
		}
		if c.AuditConfig.TABLE_COMMENT_LENGTH > 512 {
			c.AuditConfig.TABLE_COMMENT_LENGTH = 512
		}
		if c.AuditConfig.MAX_COLUMN_NAME_LENGTH > 64 {
			c.AuditConfig.MAX_COLUMN_NAME_LENGTH = 64
		}
		if c.AuditConfig.MAX_VARCHAR_LENGTH > 16383 {
			c.AuditConfig.MAX_COLUMN_NAME_LENGTH = 16383
		}
	}()
	return nil
}

func (c *Checker) Parse() error {
	// 解析SQL
	var warns []error
	var err error
	// 处理审核参数
	if err := c.CustomParams(); err != nil {
		return err
	}
	// 解析
	c.Audit, warns, err = parser.NewParse(c.Form.SqlText, c.Charset, c.Collation)
	if len(warns) > 0 {
		return fmt.Errorf("Parse Warning: %s", utils.ErrsJoin("; ", warns))
	}
	if err != nil {
		return fmt.Errorf("sql解析错误：%s", err.Error())
	}
	return nil
}

func (c *Checker) CreateViewStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// 建视图语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	for _, rule := range rules.CreateViewRules() {
		var ruleHint *controllers.RuleHint = &controllers.RuleHint{
			DB:          c.DB,
			KV:          kv,
			AuditConfig: &c.AuditConfig,
		}
		rule.RuleHint = ruleHint
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

func (c *Checker) CreateTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// 建表语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	for _, rule := range rules.CreateTableRules() {
		var ruleHint *controllers.RuleHint = &controllers.RuleHint{
			DB:          c.DB,
			KV:          kv,
			AuditConfig: &c.AuditConfig,
		}
		rule.RuleHint = ruleHint

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
	// 禁止使用ALTER TABLE...ADD CONSTRAINT...语法
	tmpCompile := regexp.MustCompile(`(?is:.*alter.*table.*add.*constraint.*)`)
	match := tmpCompile.MatchString(stmt.Text())
	if match {
		data.Level = "WARN"
		data.Summary = append(data.Summary, "禁止使用ALTER TABLE...ADD CONSTRAINT...语法")
		return data, mergeAlter
	}
	for _, rule := range rules.AlterTableRules() {
		var ruleHint *controllers.RuleHint = &controllers.RuleHint{
			DB:          c.DB,
			KV:          kv,
			AuditConfig: &c.AuditConfig,
		}
		rule.RuleHint = ruleHint
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
			// 如果IsSkipNextStep为true，跳过接下来的检查步骤
			break
		}
	}
	return data, mergeAlter
}

func (c *Checker) RenameTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// rename table语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	for _, rule := range rules.RenameTableRules() {
		var ruleHint *controllers.RuleHint = &controllers.RuleHint{
			DB:          c.DB,
			KV:          kv,
			AuditConfig: &c.AuditConfig,
		}
		rule.RuleHint = ruleHint
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

func (c *Checker) AnalyzeTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// analyze table语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	for _, rule := range rules.AnalyzeTableRules() {
		var ruleHint *controllers.RuleHint = &controllers.RuleHint{
			DB:          c.DB,
			KV:          kv,
			AuditConfig: &c.AuditConfig,
		}
		rule.RuleHint = ruleHint
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

func (c *Checker) DropTableStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerId string) ReturnData {
	// drop/truncate语句
	var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DDL", Level: "INFO"}
	for _, rule := range rules.DropTableRules() {
		var ruleHint *controllers.RuleHint = &controllers.RuleHint{
			DB:          c.DB,
			KV:          kv,
			AuditConfig: &c.AuditConfig,
		}
		rule.RuleHint = ruleHint
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
	for _, rule := range rules.DMLRules() {
		var ruleHint *controllers.RuleHint = &controllers.RuleHint{
			DB:          c.DB,
			KV:          kv,
			AuditConfig: &c.AuditConfig,
		}
		rule.RuleHint = ruleHint
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
	// 检查mysql merge操作
	var data ReturnData = ReturnData{Level: "INFO"}
	dbVersionIns := process.DbVersion{Version: kv.Get("dbVersion").(string)}
	if c.AuditConfig.ENABLE_MYSQL_MERGE_ALTER_TABLE && !dbVersionIns.IsTiDB() {
		if ok, val := utils.IsRepeat(mergeAlters); ok {
			for _, v := range val {
				data.Summary = append(data.Summary, fmt.Sprintf("[MySQL数据库]表`%s`的多条ALTER操作，请合并为一条ALTER语句", v))
			}
		}
	}
	if len(data.Summary) > 0 {
		data.Level = "WARN"
	}
	return data
}

func (c *Checker) Check() (err error, returnData []ReturnData) {
	c.InitDB()
	var mergeAlters []string // 存放alter语句中的表名

	// 记录下审计sql
	global.App.Log.WithFields(logrus.Fields{"request_id": c.Form.RequestID, "type": "App"}).Info(c.Form.SqlText)

	// 每次请求基于RequestID初始化kv cache
	kv := kv.NewKVCache(c.Form.RequestID)
	// 获取目标数据库变量
	dbVars, err := dao.GetDBVars(c.DB)
	if err != nil {
		global.App.Log.WithFields(logrus.Fields{"request_id": c.Form.RequestID, "type": "App"}).Error(err)
		return fmt.Errorf(err.Error()), returnData
	}
	for k, v := range dbVars {
		kv.Put(k, v)
	}
	c.Charset = dbVars["dbCharset"]

	// 解析SQL
	err = c.Parse()
	if err != nil {
		global.App.Log.WithFields(logrus.Fields{"request_id": c.Form.RequestID, "type": "App"}).Error(err)
		return err, returnData
	}

	// 迭代stmt
	for _, stmt := range c.Audit.TiStmt {
		// 移除SQL尾部的分号
		sqlTrim := strings.TrimSuffix(stmt.Text(), ";")
		fingerId := query.Id(query.Fingerprint(sqlTrim))
		kv.Put(fingerId, true)
		// 迭代
		switch stmt.(type) {
		case *ast.SelectStmt:
			// select语句不允许审核
			var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DML", Level: "WARN"}
			data.Summary = append(data.Summary, "发现SELECT语句，请删除SELECT语句后重新审核")
			returnData = append(returnData, data)
		case *ast.CreateTableStmt:
			returnData = append(returnData, c.CreateTableStmt(stmt, kv, fingerId))
		case *ast.CreateViewStmt:
			returnData = append(returnData, c.CreateViewStmt(stmt, kv, fingerId))
		case *ast.AlterTableStmt:
			data, mergeAlter := c.AlterTableStmt(stmt, kv, fingerId)
			mergeAlters = append(mergeAlters, mergeAlter)
			returnData = append(returnData, data)
		case *ast.DropTableStmt, *ast.TruncateTableStmt:
			returnData = append(returnData, c.DropTableStmt(stmt, kv, fingerId))
		case *ast.DeleteStmt, *ast.InsertStmt, *ast.UpdateStmt:
			returnData = append(returnData, c.DMLStmt(stmt, kv, fingerId))
		case *ast.RenameTableStmt:
			returnData = append(returnData, c.RenameTableStmt(stmt, kv, fingerId))
		case *ast.AnalyzeTableStmt:
			returnData = append(returnData, c.AnalyzeTableStmt(stmt, kv, fingerId))
		default:
			// 不允许的其他语句，有需求可以扩展
			var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "", Level: "WARN"}
			data.Summary = append(data.Summary, "不被允许的审核语句，请联系数据库管理员")
			returnData = append(returnData, data)
		}
	}
	if len(mergeAlters) > 1 {
		mergeData := c.MergeAlter(kv, mergeAlters)
		if len(mergeData.Summary) > 0 {
			returnData = append(returnData, mergeData)
		}
	}
	// 比如只传递了注释,如:#
	if len(c.Audit.TiStmt) == 0 {
		return nil, []ReturnData{}
	}
	return nil, returnData
}
