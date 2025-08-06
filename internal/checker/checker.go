package checker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/lazzyfu/gaudit/forms"
	"github.com/lazzyfu/gaudit/internal/config"
	"github.com/lazzyfu/gaudit/internal/dao"
	"github.com/lazzyfu/gaudit/internal/global"
	"github.com/lazzyfu/gaudit/internal/parser"
	"github.com/lazzyfu/gaudit/internal/process"
	"github.com/lazzyfu/gaudit/pkg/kv"
	"github.com/lazzyfu/gaudit/pkg/utils"

	query "github.com/lazzyfu/gaudit/pkg/query"

	"github.com/jinzhu/copier"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
	"github.com/sirupsen/logrus"
)

// 返回数据格式
type ReturnData struct {
	Summary      []string               `json:"summary"` // 规则摘要
	Level        string                 `json:"level"`   // 提醒级别,INFO/WARN/ERROR
	AffectedRows int                    `json:"affected_rows"`
	Type         string                 `json:"type"`
	FingerId     string                 `json:"finger_id"`
	Query        string                 `json:"query"`           // 原始SQL
	Extra        map[string]interface{} `json:"extra,omitempty"` // 扩展字段
}

// 语句处理器接口
type StmtHandler interface {
	Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData
}

// 语句处理上下文
type StmtContext struct {
	DB          *dao.DB
	Kv          *kv.KVCache[string]
	FingerId    string
	AuditConfig config.AuditConfiguration
}

// Checker主结构体
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
func (c *Checker) initAuditParams() error {
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

	// 参数边界保护
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

func (c *Checker) parser() error {
	// 解析SQL
	var warns []error
	var err error
	// 解析
	c.Audit, warns, err = parser.NewParse(c.Form.SqlText, c.Charset, c.Collation)
	if len(warns) > 0 {
		return fmt.Errorf("语法解析异常: %s", utils.ErrsJoin("; ", warns))
	}
	if err != nil {
		return fmt.Errorf("语法解析错误：%s", err.Error())
	}
	return nil
}

func (c *Checker) MergeAlter(kv *kv.KVCache[string], mergeAlters []string) ReturnData {
	var data ReturnData = ReturnData{Level: "INFO"}
	dbVersion, ok := kv.Get("dbVersion")
	if !ok {
		data.Summary = append(data.Summary, "无法获取数据库版本信息")
		return data
	}
	dbVersionIns := process.DbVersion{Version: dbVersion}
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

// 语句类型与处理器映射
func (c *Checker) stmtHandlers() map[string]StmtHandler {
	return map[string]StmtHandler{
		"*ast.SelectStmt":        &SelectHandler{},
		"*ast.CreateTableStmt":   &CreateTableHandler{},
		"*ast.CreateViewStmt":    &CreateViewHandler{},
		"*ast.AlterTableStmt":    &AlterTableHandler{},
		"*ast.DropTableStmt":     &DropTableHandler{},
		"*ast.TruncateTableStmt": &DropTableHandler{},
		"*ast.DeleteStmt":        &DMLHandler{},
		"*ast.InsertStmt":        &DMLHandler{},
		"*ast.UpdateStmt":        &DMLHandler{},
		"*ast.RenameTableStmt":   &RenameTableHandler{},
		"*ast.AnalyzeTableStmt":  &AnalyzeTableHandler{},
	}
}

// 主审核入口
func (c *Checker) Check() (err error, returnData []ReturnData) {
	c.InitDB()
	global.App.Log.WithFields(logrus.Fields{"request_id": c.RequestID}).Info(c.Form.SqlText)

	// 每次请求基于RequestID初始化kv cache
	kv := kv.NewKVCache[string](c.RequestID)
	defer kv.Delete(c.RequestID)

	// 获取目标数据库变量
	dbVars, err := dao.GetDBVars(c.DB)
	if err != nil {
		global.App.Log.WithFields(logrus.Fields{"request_id": c.RequestID}).Error(err)
		return err, returnData
	}
	for k, v := range dbVars {
		kv.Put(k, v)
	}
	c.Charset = dbVars["dbCharset"]

	// 初始化审核参数
	if err = c.initAuditParams(); err != nil {
		global.App.Log.WithFields(logrus.Fields{"request_id": c.RequestID}).Error(err)
		return
	}

	// 解析SQL
	err = c.parser()
	if err != nil {
		global.App.Log.WithFields(logrus.Fields{"request_id": c.RequestID}).Error(err)
		return
	}

	var mergeAlters []string // 存放alter语句中的表名
	handlers := c.stmtHandlers()

	// 迭代stmt
	for _, stmt := range c.Audit.TiStmt {
		// 移除SQL尾部的分号
		sqlTrim := strings.TrimSuffix(stmt.Text(), ";")
		fingerId := query.Id(query.Fingerprint(sqlTrim))
		ctx := &StmtContext{DB: c.DB, Kv: kv, FingerId: fingerId, AuditConfig: c.AuditConfig}

		stmtType := reflect.TypeOf(stmt).String()
		handler, ok := handlers[stmtType]
		if !ok {
			// 默认处理
			returnData = append(returnData, ReturnData{
				FingerId: fingerId,
				Query:    stmt.Text(),
				Type:     "",
				Level:    "WARN",
				Summary:  []string{"不被允许的审核语句，请联系数据库管理员"},
			})
			continue
		}

		// 特殊处理ALTER合并
		if stmtType == "*ast.AlterTableStmt" {
			data := handler.Handle(stmt, ctx)
			if data.Extra != nil {
				if alterTableName, ok := data.Extra["mergeAlter"].(string); ok {
					mergeAlters = append(mergeAlters, alterTableName)
				}
			}
			returnData = append(returnData, data)
		} else {
			returnData = append(returnData, handler.Handle(stmt, ctx))
		}
	}
	// 判断多条alter语句是否需要合并
	if len(mergeAlters) > 1 {
		mergeData := c.MergeAlter(kv, mergeAlters)
		if len(mergeData.Summary) > 0 {
			returnData = append(returnData, mergeData)
		}
	}
	// 只传递注释
	if len(c.Audit.TiStmt) == 0 {
		return nil, []ReturnData{}
	}

	return
}

// 以下为各类型语句处理器实现示例
type SelectHandler struct{}

func (h *SelectHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	return ReturnData{
		FingerId: ctx.FingerId,
		Query:    stmt.Text(),
		Type:     "DML",
		Level:    "WARN",
		Summary:  []string{"发现SELECT语句，请删除SELECT语句后重新审核"},
	}
}

type CreateTableHandler struct{}

func (h *CreateTableHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	st := NewStmt(ctx, stmt)
	return st.CreateTableStmt()
}

type CreateViewHandler struct{}

func (h *CreateViewHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	st := NewStmt(ctx, stmt)
	return st.CreateViewStmt()
}

type AlterTableHandler struct{}

func (h *AlterTableHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	st := NewStmt(ctx, stmt)
	data, mergeAlter := st.AlterTableStmt()
	if data.Extra == nil {
		data.Extra = make(map[string]interface{})
	}
	data.Extra["mergeAlter"] = mergeAlter
	return data
}

type DropTableHandler struct{}

func (h *DropTableHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	st := NewStmt(ctx, stmt)
	return st.DropTableStmt()
}

type DMLHandler struct{}

func (h *DMLHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	st := NewStmt(ctx, stmt)
	return st.DMLStmt()
}

type RenameTableHandler struct{}

func (h *RenameTableHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	st := NewStmt(ctx, stmt)
	return st.RenameTableStmt()
}

type AnalyzeTableHandler struct{}

func (h *AnalyzeTableHandler) Handle(stmt ast.StmtNode, ctx *StmtContext) ReturnData {
	st := NewStmt(ctx, stmt)
	return st.AnalyzeTableStmt()
}
