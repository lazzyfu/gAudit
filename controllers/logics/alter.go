/*
@Time    :   2022/07/06 10:12:00
@Author  :   xff
@Desc    :   None
*/

package logics

import (
	"fmt"
	"gAudit/config"
	"gAudit/controllers"
	"gAudit/controllers/dao"
	"gAudit/controllers/process"
	"gAudit/controllers/traverses"
	"gAudit/pkg/utils"
	"strings"
)

// LogicAlterTableIsExist
func LogicAlterTableIsExist(v *traverses.TraverseAlterTableIsExist, r *controllers.RuleHint) {
	// 检查表是否存在，如果表不存在，skip下面的检查
	r.MergeAlter = v.Table
	if err, msg := dao.DescTable(v.Table, r.DB); err != nil {
		r.Summary = append(r.Summary, msg)
		r.IsSkipNextStep = true
	}
	// 禁止审核指定的表
	for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
		if item.DB == r.DB.Database && utils.IsContain(item.Tables, v.Table) {
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核，原因：%s", r.DB.Database, v.Table, item.Reason))
			r.IsSkipNextStep = true
		}
	}
}

// LogicAlterTableTiDBMerge
func LogicAlterTableTiDBMerge(v *traverses.TraverseAlterTiDBMerge, r *controllers.RuleHint) {
	// 检查TiDBMergeAlter
	dbVersionIns := process.DbVersion{Version: r.KV.Get("dbVersion").(string)}
	if !r.AuditConfig.ENABLE_TIDB_MERGE_ALTER_TABLE && dbVersionIns.IsTiDB() {
		if v.SpecsLen > 1 {
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`的多个操作，请拆分为多条ALTER语句(TiDB不支持在单个ALTER TABLE语句中进行多个更改)", v.Table))
		}
	}
}

// LogicAlterTableDropColsOrIndexes
func LogicAlterTableDropColsOrIndexes(v *traverses.TraverseAlterTableDropColsOrIndexes, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}

	if len(v.Cols) > 0 {
		if !r.AuditConfig.ENABLE_DROP_COLS {
			// 不允许drop列
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不允许DROP列", v.Table))
		} else {
			// 检查drop的列是否存在
			for _, col := range v.Cols {
				if !utils.IsContain(vAudit.Cols, col) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`DROP的列`%s`不存在", v.Table, col))
				}
			}
		}
		if !r.AuditConfig.ENABLE_DROP_PRIMARYKEY {
			// 不允许drop主键
			for _, pri := range vAudit.PrimaryKeys {
				if utils.IsContain(v.Cols, pri) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不允许DROP主键`%s`", v.Table, pri))
				}
			}
		}
	}
	if len(vAudit.Indexes) > 0 {
		if !r.AuditConfig.ENABLE_DROP_INDEXES {
			// 不允许drop索引
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不允许DROP索引", v.Table))
		} else {
			// 检查drop的索引是否存在
			for _, index := range v.Indexes {
				if !utils.IsContain(vAudit.Indexes, index) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`DROP的索引`%s`不存在", v.Table, index))
				}
			}
		}
	}
}

// LogicAlterTableDropTiDBColWithCoveredIndex
func LogicAlterTableDropTiDBColWithCoveredIndex(v *traverses.TraverseAlterTableDropTiDBColWithCoveredIndex, r *controllers.RuleHint) {
	// TiDB目前不支持删除主键列或组合索引相关列。
	dbVersionIns := process.DbVersion{Version: r.KV.Get("dbVersion").(string)}
	if !dbVersionIns.IsTiDB() {
		return
	}
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableRedundantIndexes{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}

	for _, col := range v.Cols {
		for _, item := range vAudit.Redundant.IndexesCols {
			if len(item.Cols) > 1 {
				if utils.IsContain(item.Cols, col) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`DROP列`%s`操作失败，无法删除包含组合索引的列，当前列已经被组合索引%s(%s)覆盖【TiDB目前不支持删除主键列或组合索引相关列，请先删除复合索引`%s`】", v.Table, col, item.Index, strings.Join(item.Cols, ","), item.Index))
				}
			}
		}
	}

}

// LogicAlterTableOptions
func LogicAlterTableOptions(v *traverses.TraverseAlterTableOptions, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table
	v.Type = "alter"
	v.TableOptions.AuditConfig = r.AuditConfig

	// 行格式
	var rowFormat string = v.RowFormat
	if v.RowFormat == "DEFAULT" {
		rowFormat = r.KV.Get("innodbDefaultRowFormat").(string)
	}
	v.TableOptions.RowFormat = rowFormat

	fns := []func() error{v.CheckTableEngine, v.CheckTableComment, v.CheckTableCharset, v.CheckInnoDBRowFormat}
	for _, fn := range fns {
		if err := fn(); err != nil {
			r.Summary = append(r.Summary, err.Error())
		}
	}
}

// LogicAlterTableColCharset
func LogicAlterTableColCharset(v *traverses.TraverseAlterTableColCharset, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 列字符集检查
	if r.AuditConfig.CHECK_COLUMN_CHARSET {
		if len(v.Cols) > 0 {
			if err := v.CheckColumn(); err != nil {
				r.Summary = append(r.Summary, err.Error())
			}
		}
	}
}

// LogicAlterTableAddColAfter
func LogicAlterTableAddColAfter(v *traverses.TraverseAlterTableAddColAfter, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}

	// 将add的列和原始表的列放到一起
	v.Cols = append(v.Cols, vAudit.Cols...)

	// 检查AFTER的列是否存在
	for _, pCol := range v.PositionCols {
		if !utils.IsContain(v.Cols, pCol) {
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`语句中AFTER指定的列`%s`不存在", v.Table, pCol))
		}
	}
}

// LogicAlterTableAddColOptions
func LogicAlterTableAddColOptions(v *traverses.TraverseAlterTableAddColOptions, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	for _, col := range v.Cols {
		col.AuditConfig = r.AuditConfig
		fns := []func() error{
			col.CheckColumnLength,
			col.CheckColumnIdentifer,
			col.CheckColumnIdentiferKeyword,
			col.CheckColumnComment,
			col.CheckColumnCharToVarchar,
			col.CheckColumnMaxVarcharLength,
			col.CheckColumnFloatDouble,
			col.CheckColumnNotAllowedType,
			col.CheckColumnNotNull,
			col.CheckColumnDefaultValue,
		}
		for _, fn := range fns {
			if err := fn(); err != nil {
				r.Summary = append(r.Summary, err.Error())
			}
		}
	}
}

// LogicAlterTableAddPrimaryKey
func LogicAlterTableAddPrimaryKey(v *traverses.TraverseAlterTableAddPrimaryKey, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}

	if len(vAudit.PrimaryKeys) > 0 && len(v.Cols) > 0 {
		var newPrimaryKeys []string
		for _, col := range v.Cols {
			newPrimaryKeys = append(newPrimaryKeys, fmt.Sprintf("`%s`", col))
		}
		r.Summary = append(r.Summary, fmt.Sprintf("表`%s`已经存在主键`%s`，增加主键%+s失败", v.Table, strings.Join(vAudit.PrimaryKeys, ","), strings.Join(newPrimaryKeys, ",")))
	}
}

// LogicAlterTableAddColRepeatDefine
func LogicAlterTableAddColRepeatDefine(v *traverses.TraverseAlterTableAddColRepeatDefine, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 查找重复的列名
	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableColsRepeatDefine{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	v.Cols = append(v.Cols, vAudit.Cols...)

	if ok, data := utils.IsRepeat(v.Cols); ok {
		r.Summary = append(r.Summary, fmt.Sprintf("发现重复的列名`%s`[表`%s`]", strings.Join(data, ","), v.Table))
	}
}

// LogicAlterTableAddIndexPrefix
func LogicAlterTableAddIndexPrefix(v *traverses.TraverseAlterTableAddIndexPrefix, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 检查唯一索引前缀、如唯一索引必须以uniq_为前缀
	var indexPrefixCheck process.IndexPrefix = v.Prefix
	indexPrefixCheck.AuditConfig = r.AuditConfig
	if r.AuditConfig.CHECK_UNIQ_INDEX_PREFIX {
		if err := indexPrefixCheck.CheckUniquePrefix(); err != nil {
			r.Summary = append(r.Summary, err.Error())
		}
	}
	// 检查二级索引前缀、如二级索引必须以idx_为前缀
	if r.AuditConfig.CHECK_SECONDARY_INDEX_PREFIX {
		if err := indexPrefixCheck.CheckSecondaryPrefix(); err != nil {
			r.Summary = append(r.Summary, err.Error())
		}
	}
	// 检查全文索引前缀、如全文索引必须以full_为前缀
	if r.AuditConfig.CHECK_FULLTEXT_INDEX_PREFIX {
		if err := indexPrefixCheck.CheckFulltextPrefix(); err != nil {
			r.Summary = append(r.Summary, err.Error())
		}
	}
}

// LogicAlterTableAddIndexCount
func LogicAlterTableAddIndexCount(v *traverses.TraverseAlterTableAddIndexCount, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableIndexesCount{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	v.Number.Number += vAudit.Number.Number
	// 检查二级索引数量
	var indexNumberCheck process.IndexNumber = v.Number
	indexNumberCheck.AuditConfig = r.AuditConfig
	if err := indexNumberCheck.CheckSecondaryIndexesNum(); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
	if err := indexNumberCheck.CheckPrimaryKeyColsNum(); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
}

// LogicAlterTableAddConstraint
func LogicAlterTableAddConstraint(v *traverses.TraverseAlterTableAddConstraint, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table
	if !r.AuditConfig.ENABLE_FOREIGN_KEY && v.IsForeignKey {
		// 禁止使用外键
		r.Summary = append(r.Summary, fmt.Sprintf("表`%s`禁止定义外键", v.Table))
	}
}

// LogicAlterTableAddIndexRepeatDefine
func LogicAlterTableAddIndexRepeatDefine(v *traverses.TraverseAlterTableAddIndexRepeatDefine, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableIndexesRepeatDefine{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	v.Indexes = append(v.Indexes, vAudit.Indexes...)
	if ok, data := utils.IsRepeat(v.Indexes); ok {
		r.Summary = append(r.Summary, fmt.Sprintf("发现重复的索引`%s`[表`%s`]", strings.Join(data, ","), v.Table))
	}
}

// LogicAlterTableRedundantIndexes
func LogicAlterTableRedundantIndexes(v *traverses.TraverseAlterTableRedundantIndexes, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}

	if r.AuditConfig.ENABLE_REDUNDANT_INDEX {
		return
	}

	r.MergeAlter = v.Table

	// 检查索引,建索引时,指定的列必须存在、索引中的列,不能重复、索引名不能重复
	// 不能有重复的索引,包括(索引名不同,字段相同；冗余索引,如(a),(a,b))
	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableRedundantIndexes{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	v.Redundant.Cols = vAudit.Redundant.Cols
	// 用于检查alter table xxx add `col1` xxx,add index idx_col1(`col1`)
	v.Redundant.Cols = append(v.Redundant.Cols, v.AddCols...)
	// 用于检查alter table xxx drop `col2`,add index idx_col2(`col2`);
	v.Redundant.Cols = utils.RemoveElements(v.Redundant.Cols, v.DropCols)
	v.Redundant.Indexes = append(v.Redundant.Indexes, vAudit.Redundant.Indexes...)
	v.Redundant.IndexesCols = append(v.Redundant.IndexesCols, vAudit.Redundant.IndexesCols...)
	var redundantIndexCheck process.RedundantIndex = v.Redundant
	if err := redundantIndexCheck.CheckRepeatCols(); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
	if err := redundantIndexCheck.CheckRepeatColsWithDiffIndexes(); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
	if err := redundantIndexCheck.CheckRedundantColsWithDiffIndexes(); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
}

// LogicAlterTableDisabledIndexes
func LogicAlterTableDisabledIndexes(v *traverses.TraverseAlterTableDisabledIndexes, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableDisabledIndexes{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	v.DisabledIndexes.Cols = append(v.DisabledIndexes.Cols, vAudit.DisabledIndexes.Cols...)
	v.DisabledIndexes.IndexesCols = append(v.DisabledIndexes.IndexesCols, vAudit.DisabledIndexes.IndexesCols...)

	// BLOB/TEXT类型不能设置为索引
	var indexTypesCheck process.DisabledIndexes = v.DisabledIndexes
	if err := indexTypesCheck.Check(); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
}

// LogicAlterTableModifyColOptions
func LogicAlterTableModifyColOptions(v *traverses.TraverseAlterTableModifyColOptions, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableColsOptions{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	var vCols []string
	for _, vCol := range vAudit.Cols {
		vCols = append(vCols, vCol.Column)
	}
	// 检查modify的列是否存在
	for _, col := range v.Cols {
		if !utils.IsContain(vCols, col.Column) {
			r.Summary = append(r.Summary, fmt.Sprintf("列`%s`不存在[表`%s`]", col.Column, v.Table))
		}
	}
	// 检查modify的列是否进行列类型变更
	for _, col := range v.Cols {
		for _, vCol := range vAudit.Cols {
			if err := process.CheckColsTypeChanged(col, vCol, r.AuditConfig, r.KV, "modify", v.Table); err != nil {
				r.Summary = append(r.Summary, err.Error())
			}
		}
	}
	// 检查列
	for _, col := range v.Cols {
		col.AuditConfig = r.AuditConfig
		fns := []func() error{
			col.CheckColumnComment,
			col.CheckColumnCharToVarchar,
			col.CheckColumnMaxVarcharLength,
			col.CheckColumnNotAllowedType,
			col.CheckColumnNotNull,
			col.CheckColumnDefaultValue,
		}
		for _, fn := range fns {
			if err := fn(); err != nil {
				r.Summary = append(r.Summary, err.Error())
			}
		}
	}
}

// LogicAlterTableChangeColOptions
func LogicAlterTableChangeColOptions(v *traverses.TraverseAlterTableChangeColOptions, r *controllers.RuleHint) {
	/*
		change操作的2种用法
		修改列的类型信息
			> ALTER TABLE 【表名字】 CHANGE 【列名称】【列名称】 BIGINT NOT NULL  COMMENT '注释说明'
		修改列名+修改列类型信息
			> ALTER TABLE 【表名字】 CHANGE 【列名称】【新列名称】 BIGINT NOT NULL  COMMENT '注释说明'
	*/
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table
	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableColsOptions{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	var vCols []string
	for _, vCol := range vAudit.Cols {
		vCols = append(vCols, vCol.Column)
	}
	// 判断change操作是否为修改列名操作
	for _, col := range v.Cols {
		if col.Column != col.OldColumn {
			// 不允许change列名操作
			if !r.AuditConfig.ENABLE_COLUMN_CHANGE_COLUMN_NAME && len(v.Cols) > 0 {
				r.Summary = append(r.Summary, fmt.Sprintf("禁止CHANGE修改列名操作(`%s` -> `%s`)[表`%s`]", col.OldColumn, col.Column, v.Table))
				return
			}
			// 允许change列名操作,检查change的列是否存在
			if !utils.IsContain(vCols, col.OldColumn) {
				r.Summary = append(r.Summary, fmt.Sprintf("原列`%s`不存在[表`%s`]", col.OldColumn, v.Table))
			}
			if utils.IsContain(vCols, col.Column) {
				r.Summary = append(r.Summary, fmt.Sprintf("新列`%s`已经存在[表`%s`]", col.Column, v.Table))
			}
		} else {
			// 允许change列名操作,检查change的列是否存在
			if !utils.IsContain(vCols, col.OldColumn) {
				r.Summary = append(r.Summary, fmt.Sprintf("原列`%s`不存在[表`%s`]", col.OldColumn, v.Table))
			}
		}
	}

	// 检查change的列是否进行列类型变更
	for _, col := range v.Cols {
		for _, vCol := range vAudit.Cols {
			if col.OldColumn == vCol.Column {
				if err := process.CheckColsTypeChanged(col, vCol, r.AuditConfig, r.KV, "change", v.Table); err != nil {
					r.Summary = append(r.Summary, err.Error())
				}
			}
		}
	}

	// 检查列
	for _, col := range v.Cols {
		col.AuditConfig = r.AuditConfig
		fns := []func() error{
			col.CheckColumnComment,
			col.CheckColumnCharToVarchar,
			col.CheckColumnMaxVarcharLength,
			col.CheckColumnFloatDouble,
			col.CheckColumnNotAllowedType,
			col.CheckColumnNotNull,
			col.CheckColumnDefaultValue,
		}
		for _, fn := range fns {
			if err := fn(); err != nil {
				r.Summary = append(r.Summary, err.Error())
			}
		}
	}
}

// LogicAlterTableRenameIndex
func LogicAlterTableRenameIndex(v *traverses.TraverseAlterTableRenameIndex, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	if !r.AuditConfig.ENABLE_INDEX_RENAME {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许RENAME INDEX操作[表`%s`]", v.Table))
		return
	}
	// 判断是否重复定义
	var oldIndexes, newIndexes []string
	for _, item := range v.Indexes {
		oldIndexes = append(oldIndexes, item.OldIndex)
		newIndexes = append(newIndexes, item.NewIndex)
	}
	if ok, data := utils.IsRepeat(oldIndexes); ok {
		r.Summary = append(r.Summary, fmt.Sprintf("发现操作重复的索引`%s`[表`%s`]", strings.Join(data, ","), v.Table))
	}
	if ok, data := utils.IsRepeat(newIndexes); ok {
		r.Summary = append(r.Summary, fmt.Sprintf("发现操作重复的索引`%s`[表`%s`]", strings.Join(data, ","), v.Table))
	}
	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	// 判断表是否存在
	if v.Table != vAudit.Table {
		r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不存在", v.Table))
	}
	for _, item := range v.Indexes {
		// 判断old_index_name是否存在
		if !utils.IsContain(vAudit.Indexes, item.OldIndex) {
			r.Summary = append(r.Summary, fmt.Sprintf("索引`%s`不存在[表`%s`]", item.OldIndex, v.Table))
		}
		// 判断new_index_name是否存在
		if utils.IsContain(vAudit.Indexes, item.NewIndex) {
			r.Summary = append(r.Summary, fmt.Sprintf("新的索引`%s`已存在[表`%s`]", item.NewIndex, v.Table))
		}
		// 检查索引名合法性
		if r.AuditConfig.CHECK_IDENTIFIER {
			if ok := utils.IsMatchPattern(utils.NamePattern, item.NewIndex); !ok {
				r.Summary = append(r.Summary, fmt.Sprintf("索引`%s`命名不符合要求，仅允许匹配正则`%s`[表`%s`]", item.NewIndex, utils.NamePattern, v.Table))
			}
		}
	}
}

// LogicAlterTableRenameTblName
func LogicAlterTableRenameTblName(v *traverses.TraverseAlterTableRenameTblName, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table
	if !r.AuditConfig.ENABLE_RENAME_TABLE_NAME {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许RENAME表名[表`%s`]", v.Table))
		return
	}
	// 判断新表是否存在
	if err, msg := dao.DescTable(v.NewTblName, r.DB); err == nil {
		r.Summary = append(r.Summary, msg)
		return
	}
}

// LogicAlterTableInnodbLargePrefix
func LogicAlterTableInnodbLargePrefix(v *traverses.TraverseAlterTableInnodbLargePrefix, r *controllers.RuleHint) {
	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.LargePrefix.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableColsTp{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	// 将提前到的字段类型复制给索引字段结构体
	var tmpLargePrefix process.LargePrefix
	tmpLargePrefix.Table = v.LargePrefix.Table
	tmpLargePrefix.Charset = vAudit.Charset
	for _, i := range v.LargePrefix.LargePrefixIndexColsMaps {
		var tmpKeys []process.LargePrefixIndexPartSpecification = i.Keys
		for index, ii := range i.Keys {
			for _, jj := range vAudit.Cols {
				if strings.EqualFold(jj.Column, ii.Column) {
					tmpKeys[index].Tp = jj.Tp
					tmpKeys[index].Flen = jj.Flen
					tmpKeys[index].Decimal = jj.Decimal
					tmpKeys[index].Charset = jj.Charset
				}
			}
		}
		tmpLargePrefix.LargePrefixIndexColsMaps = append(tmpLargePrefix.LargePrefixIndexColsMaps, process.LargePrefixIndexColsMap{Name: i.Name, Keys: tmpKeys})
	}

	var largePrefix process.LargePrefix = tmpLargePrefix
	if err := largePrefix.Check(r.KV); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
}

// LogicAlterTableInnoDBRowSize
func LogicAlterTableInnoDBRowSize(v *traverses.TraverseAlterTableInnoDBRowSize, r *controllers.RuleHint) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := dao.ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAudit := &traverses.TraverseCreateTableInnoDBRowSize{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAudit)
	}
	// 拷贝，如果Column不存在append，Column存在，重新赋值
	for _, v := range v.ColsMaps {
		if index, ok := func(v process.PartSpecification) (int, bool) {
			for i, vv := range vAudit.ColsMaps {
				if strings.EqualFold(v.Column, vv.Column) {
					return i, true
				}
			}
			return 0, false
		}(v); !ok {
			vAudit.InnoDBRowSize.ColsMaps = append(vAudit.ColsMaps, v)
		} else {
			vAudit.ColsMaps[index] = v
		}

	}
	var rowSizeTooLarge process.InnoDBRowSize = vAudit.InnoDBRowSize
	if err := rowSizeTooLarge.Check(r.KV); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
}
