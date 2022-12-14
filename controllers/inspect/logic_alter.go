/*
@Time    :   2022/07/06 10:12:00
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/controllers/process"
	"strings"
)

// LogicAlterTableIsExist
func LogicAlterTableIsExist(v *TraverseAlterTableIsExist, r *Rule) {
	// 检查表是否存在，如果表不存在，skip下面的检查
	r.MergeAlter = v.Table
	if err, msg := DescTable(v.Table, r.DB); err != nil {
		r.Summary = append(r.Summary, msg)
		r.IsSkipNextStep = true
	}
	// 禁止审核指定的表
	if len(r.AuditConfig.DISABLE_AUDIT_DDL_TABLES) > 0 {
		for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
			if item.DB == r.DB.Database && utils.IsContain(item.Tables, v.Table) {
				r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核,原因: %s", r.DB.Database, v.Table, item.Reason))
				r.IsSkipNextStep = true
			}
		}
	}
}

// LogicAlterTableTiDBMerge
func LogicAlterTableTiDBMerge(v *TraverseAlterTiDBMerge, r *Rule) {
	// 检查TiDBMergeAlter
	dbVersionIns := process.DbVersion{Version: r.KV.Get("dbVersion").(string)}
	if !r.AuditConfig.ENABLE_TIDB_MERGE_ALTER_TABLE && dbVersionIns.IsTiDB() {
		if v.SpecsLen > 1 {
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`的多个操作,请拆分为多条ALTER语句(TiDB不支持在单个ALTER TABLE语句中进行多个更改)", v.Table))
		}
	}
}

// LogicAlterTableDropColsOrIndexes
func LogicAlterTableDropColsOrIndexes(v *TraverseAlterTableDropColsOrIndexes, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}

	if len(v.Cols) > 0 {
		if !r.AuditConfig.ENABLE_DROP_COLS {
			// 不允许drop列
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不允许DROP列", v.Table))
		} else {
			// 检查drop的列是否存在
			for _, col := range v.Cols {
				if !utils.IsContain(vAduit.Cols, col) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`DROP的列`%s`不存在", v.Table, col))
				}
			}
		}
		if !r.AuditConfig.ENABLE_DROP_PRIMARYKEY {
			// 不允许drop主键
			for _, pri := range vAduit.PrimaryKeys {
				if utils.IsContain(v.Cols, pri) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不允许DROP主键`%s`", v.Table, pri))
				}
			}
		}
	}
	if len(vAduit.Indexes) > 0 {
		if !r.AuditConfig.ENABLE_DROP_INDEXES {
			// 不允许drop索引
			r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不允许DROP索引", v.Table))
		} else {
			// 检查drop的索引是否存在
			for _, index := range v.Indexes {
				if !utils.IsContain(vAduit.Indexes, index) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`DROP的索引`%s`不存在", v.Table, index))
				}
			}
		}
	}
}

// LogicAlterTableDropTiDBColWithCoveredIndex
func LogicAlterTableDropTiDBColWithCoveredIndex(v *TraverseAlterTableDropTiDBColWithCoveredIndex, r *Rule) {
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
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableRedundantIndexes{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}

	for _, col := range v.Cols {
		for _, item := range vAduit.Redundant.IndexesCols {
			if len(item.Cols) > 1 {
				if utils.IsContain(item.Cols, col) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`DROP列`%s`操作失败,无法删除包含组合索引的列,当前列已经被组合索引%s(%s)覆盖【TiDB目前不支持删除主键列或组合索引相关列,请先删除复合索引`%s`】", v.Table, col, item.Index, strings.Join(item.Cols, ","), item.Index))
				}
			}
		}
	}

}

// LogicAlterTableOptions
func LogicAlterTableOptions(v *TraverseAlterTableOptions, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table
	v.Type = "alter"
	v.TableOptions.AuditConfig = r.AuditConfig
	fns := []func() error{v.CheckTableEngine, v.CheckTableComment, v.CheckTableCharset}
	for _, fn := range fns {
		if err := fn(); err != nil {
			r.Summary = append(r.Summary, err.Error())
		}
	}
}

// LogicAlterTableColCharset
func LogicAlterTableColCharset(v *TraverseAlterTableColCharset, r *Rule) {
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

// LogicAlterTableAddColOptions
func LogicAlterTableAddColOptions(v *TraverseAlterTableAddColOptions, r *Rule) {
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
func LogicAlterTableAddPrimaryKey(v *TraverseAlterTableAddPrimaryKey, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}

	if len(vAduit.PrimaryKeys) > 0 && len(v.Cols) > 0 {
		var newPrimaryKeys []string
		for _, col := range v.Cols {
			newPrimaryKeys = append(newPrimaryKeys, fmt.Sprintf("`%s`", col))
		}
		r.Summary = append(r.Summary, fmt.Sprintf("表`%s`已经存在主键`%s`,增加主键%+s失败", v.Table, strings.Join(vAduit.PrimaryKeys, ","), strings.Join(newPrimaryKeys, ",")))
	}
}

// LogicAlterTableAddColRepeatDefine
func LogicAlterTableAddColRepeatDefine(v *TraverseAlterTableAddColRepeatDefine, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 查找重复的列名
	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableColsRepeatDefine{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	v.Cols = append(v.Cols, vAduit.Cols...)

	if ok, data := utils.IsRepeat(v.Cols); ok {
		r.Summary = append(r.Summary, fmt.Sprintf("发现重复的列名`%s`[表`%s`]", strings.Join(data, ","), v.Table))
	}
}

// LogicAlterTableAddIndexPrefix
func LogicAlterTableAddIndexPrefix(v *TraverseAlterTableAddIndexPrefix, r *Rule) {
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
func LogicAlterTableAddIndexCount(v *TraverseAlterTableAddIndexCount, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableIndexesCount{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	v.Number.Number += vAduit.Number.Number
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

// LogicAlterTableAddIndexRepeatDefine
func LogicAlterTableAddIndexRepeatDefine(v *TraverseAlterTableAddIndexRepeatDefine, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableIndexesRepeatDefine{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	v.Indexes = append(v.Indexes, vAduit.Indexes...)
	if ok, data := utils.IsRepeat(v.Indexes); ok {
		r.Summary = append(r.Summary, fmt.Sprintf("发现重复的索引`%s`[表`%s`]", strings.Join(data, ","), v.Table))
	}
}

// LogicAlterTableRedundantIndexes
func LogicAlterTableRedundantIndexes(v *TraverseAlterTableRedundantIndexes, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 检查索引,建索引时,指定的列必须存在、索引中的列,不能重复、索引名不能重复
	// 不能有重复的索引,包括(索引名不同,字段相同；冗余索引,如(a),(a,b))
	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableRedundantIndexes{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	v.Redundant.Cols = vAduit.Redundant.Cols
	v.Redundant.Indexes = append(v.Redundant.Indexes, vAduit.Redundant.Indexes...)
	v.Redundant.IndexesCols = append(v.Redundant.IndexesCols, vAduit.Redundant.IndexesCols...)
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
func LogicAlterTableDisabledIndexes(v *TraverseAlterTableDisabledIndexes, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableDisabledIndexes{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	v.DisabledIndexes.Cols = append(v.DisabledIndexes.Cols, vAduit.DisabledIndexes.Cols...)
	v.DisabledIndexes.IndexesCols = append(v.DisabledIndexes.IndexesCols, vAduit.DisabledIndexes.IndexesCols...)

	// BLOB/TEXT类型不能设置为索引
	var indexTypesCheck process.DisabledIndexes = v.DisabledIndexes
	if err := indexTypesCheck.Check(); err != nil {
		r.Summary = append(r.Summary, err.Error())
	}
}

// LogicAlterTableModifyColOptions
func LogicAlterTableModifyColOptions(v *TraverseAlterTableModifyColOptions, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table

	// 获取db表结构
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableColsOptions{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	var vCols []string
	for _, vCol := range vAduit.Cols {
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
		for _, vCol := range vAduit.Cols {
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
func LogicAlterTableChangeColOptions(v *TraverseAlterTableChangeColOptions, r *Rule) {
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
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseCreateTableColsOptions{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	var vCols []string
	for _, vCol := range vAduit.Cols {
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
		for _, vCol := range vAduit.Cols {
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
func LogicAlterTableRenameIndex(v *TraverseAlterTableRenameIndex, r *Rule) {
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
	audit, err := ShowCreateTable(v.Table, r.DB, r.KV)
	if err != nil {
		r.Summary = append(r.Summary, err.Error())
		return
	}
	// 解析获取的db表结构
	vAduit := &TraverseAlterTableShowCreateTableGetCols{}
	switch audit := audit.(type) {
	case *config.Audit:
		(audit.TiStmt[0]).Accept(vAduit)
	}
	// 判断表是否存在
	if v.Table != vAduit.Table {
		r.Summary = append(r.Summary, fmt.Sprintf("表`%s`不存在", v.Table))
	}
	for _, item := range v.Indexes {
		// 判断old_index_name是否存在
		if !utils.IsContain(vAduit.Indexes, item.OldIndex) {
			r.Summary = append(r.Summary, fmt.Sprintf("索引`%s`不存在[表`%s`]", item.OldIndex, v.Table))
		}
		// 判断new_index_name是否存在
		if utils.IsContain(vAduit.Indexes, item.NewIndex) {
			r.Summary = append(r.Summary, fmt.Sprintf("新的索引`%s`已存在[表`%s`]", item.NewIndex, v.Table))
		}
		// 检查索引名合法性
		if r.AuditConfig.CHECK_IDENTIFIER {
			if ok := utils.IsMatchPattern(utils.NamePattern, item.NewIndex); !ok {
				r.Summary = append(r.Summary, fmt.Sprintf("索引`%s`命名不符合要求,仅允许匹配正则`%s`[表`%s`]", item.NewIndex, utils.NamePattern, v.Table))
			}
		}
	}
}

// LogicAlterTableRenameTblName
func LogicAlterTableRenameTblName(v *TraverseAlterTableRenameTblName, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	r.MergeAlter = v.Table
	if !r.AuditConfig.ENABLE_RENAME_TABLE_NAME {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许RENAME表名[表`%s`]", v.Table))
		return
	}
	// 判断新表是否存在
	if err, msg := DescTable(v.NewTblName, r.DB); err == nil {
		r.Summary = append(r.Summary, msg)
		return
	}
}
