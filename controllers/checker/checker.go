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
	"gAudit/controllers/dao"
	"gAudit/controllers/parser"
	"gAudit/controllers/process"
	"gAudit/forms"
	"gAudit/global"
	"gAudit/pkg/kv"
	"gAudit/pkg/utils"
	"reflect"
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
	Form        *forms.SyntaxAuditForm
	RequestID   string
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

	// 记录下审计的SQL语句
	global.App.Log.WithFields(logrus.Fields{"request_id": c.RequestID}).Info(c.Form.SqlText)

	// 每次请求基于RequestID初始化kv cache
	kv := kv.NewKVCache(c.RequestID)
	defer kv.Delete(c.RequestID)

	// 获取目标数据库变量
	dbVars, err := dao.GetDBVars(c.DB)
	if err != nil {
		global.App.Log.WithFields(logrus.Fields{"request_id": c.RequestID}).Error(err)
		return fmt.Errorf(err.Error()), returnData
	}
	for k, v := range dbVars {
		kv.Put(k, v)
	}
	c.Charset = dbVars["dbCharset"]

	// 解析SQL
	err = c.Parse()
	if err != nil {
		global.App.Log.WithFields(logrus.Fields{"request_id": c.RequestID}).Error(err)
		return err, returnData
	}

	// 迭代stmt
	for _, stmt := range c.Audit.TiStmt {
		// 移除SQL尾部的分号
		sqlTrim := strings.TrimSuffix(stmt.Text(), ";")
		fingerId := query.Id(query.Fingerprint(sqlTrim))
		kv.Put(fingerId, true)
		// 迭代
		st := Stmt{c.DB, c.AuditConfig}

		switch stmt.(type) {
		case *ast.SelectStmt:
			// select语句不允许审核
			var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "DML", Level: "WARN"}
			data.Summary = append(data.Summary, "发现SELECT语句，请删除SELECT语句后重新审核")
			returnData = append(returnData, data)
		case *ast.CreateTableStmt:
			returnData = append(returnData, st.CreateTableStmt(stmt, kv, fingerId))
		case *ast.CreateViewStmt:
			returnData = append(returnData, st.CreateViewStmt(stmt, kv, fingerId))
		case *ast.AlterTableStmt:
			data, mergeAlter := st.AlterTableStmt(stmt, kv, fingerId)
			mergeAlters = append(mergeAlters, mergeAlter)
			returnData = append(returnData, data)
		case *ast.DropTableStmt, *ast.TruncateTableStmt:
			returnData = append(returnData, st.DropTableStmt(stmt, kv, fingerId))
		case *ast.DeleteStmt, *ast.InsertStmt, *ast.UpdateStmt:
			returnData = append(returnData, st.DMLStmt(stmt, kv, fingerId))
		case *ast.RenameTableStmt:
			returnData = append(returnData, st.RenameTableStmt(stmt, kv, fingerId))
		case *ast.AnalyzeTableStmt:
			returnData = append(returnData, st.AnalyzeTableStmt(stmt, kv, fingerId))
		default:
			// 不允许的其他语句，有需求可以扩展
			var data ReturnData = ReturnData{FingerId: fingerId, Query: stmt.Text(), Type: "", Level: "WARN"}
			data.Summary = append(data.Summary, "不被允许的审核语句，请联系数据库管理员")
			returnData = append(returnData, data)
		}
	}
	// 判断多条alter语句是否需要合并
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

	return
}
