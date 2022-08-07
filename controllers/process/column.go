/*
@Time    :   2022/07/06 10:12:48
@Author  :   zongfei.fu
@Desc    :   None
*/

package process

import (
	"fmt"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/global"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pingcap/parser/mysql"
)

// 检查列的属性
type ColOptions struct {
	Table           string      // 表名
	OldColumn       string      // 旧列, CHANGE [COLUMN] old_col_name new_col_name中的old_col_name
	Column          string      // 列名
	Tp              byte        // 列类型
	Flen            int         // 类型长度
	NotNullFlag     bool        // 列是否NOT NULL
	HasDefaultValue bool        // 列是否有默认值
	DefaultValue    interface{} // 列的默认值
	DefaultIsNull   bool        // 列默认值是否为NULL
	HasComment      bool        // 是否有注释
}

// 检查列名长度
func (c *ColOptions) CheckColumnLength() error {
	if utf8.RuneCountInString(c.Column) > global.App.AuditConfig.MAX_COLUMN_NAME_LENGTH {
		return fmt.Errorf("列`%s`字符数超出限制,最大字符限制为%d[表`%s`]", c.Column, global.App.AuditConfig.MAX_COLUMN_NAME_LENGTH, c.Table)
	}
	return nil
}

// 检查列名合法性
func (c *ColOptions) CheckColumnIdentifer() error {
	if global.App.AuditConfig.CHECK_IDENTIFIER {
		if ok := utils.IsMatchPattern(utils.NamePattern, c.Column); !ok {
			return fmt.Errorf("列`%s`命名不符合要求[表`%s`]", c.Column, c.Table)
		}
	}
	return nil
}

// 检查列名是否为关键字
func (c *ColOptions) CheckColumnIdentiferKeyword() error {
	if global.App.AuditConfig.CHECK_IDENTIFER_KEYWORD {
		if _, ok := Keywords[strings.ToUpper(c.Column)]; ok {
			return fmt.Errorf("列`%s`命名不允许使用关键字[表`%s`]", c.Column, c.Table)
		}
	}
	return nil
}

// 检查列注释
func (c *ColOptions) CheckColumnComment() error {
	if global.App.AuditConfig.CHECK_COLUMN_COMMENT && !c.HasComment {
		return fmt.Errorf("列`%s`必须要有注释[表`%s`]", c.Column, c.Table)
	}
	return nil
}

// char建议转换为varchar
func (c *ColOptions) CheckColumnCharToVarchar() error {
	if global.App.AuditConfig.COLUMN_MAX_CHAR_LENGTH < c.Flen && c.Tp == mysql.TypeString {
		return fmt.Errorf("列`%s`推荐设置为varchar(%d)[表`%s`]", c.Column, c.Flen, c.Table)
	}
	return nil
}

// 最大允许定义的varchar长度
func (c *ColOptions) CheckColumnMaxVarcharLength() error {
	if global.App.AuditConfig.MAX_VARCHAR_LENGTH < c.Flen && c.Tp == mysql.TypeVarchar {
		return fmt.Errorf("列`%s`最大允许定义的varchar长度为%d,当前varchar长度为%d[表`%s`]", c.Column, global.App.AuditConfig.MAX_VARCHAR_LENGTH, c.Flen, c.Table)
	}
	return nil
}

// 将float/double转成int/bigint/decimal等
func (c *ColOptions) CheckColumnFloatDouble() error {
	if global.App.AuditConfig.CHECK_COLUMN_FLOAT_DOUBLE {
		if c.Tp == mysql.TypeFloat || c.Tp == mysql.TypeDouble {
			return fmt.Errorf("列`%s`的类型为float或double,建议转换为int/bigint/decimal类型[表`%s`]", c.Column, c.Table)
		}
	}
	return nil
}

// 列不允许定义的类型
func (c *ColOptions) CheckColumnNotAllowedType() error {
	if !global.App.AuditConfig.ENABLE_COLUMN_JSON_TYPE && c.Tp == mysql.TypeJSON {
		return fmt.Errorf("列`%s`不允许定义JSON类型[表`%s`]", c.Column, c.Table)
	}
	if !global.App.AuditConfig.ENABLE_COLUMN_BLOB_TYPE && (c.Tp == mysql.TypeTinyBlob || c.Tp == mysql.TypeMediumBlob || c.Tp == mysql.TypeLongBlob || c.Tp == mysql.TypeBlob) {
		return fmt.Errorf("列`%s`不允许定义BLOB/TEXT类型[表`%s`]", c.Table, c.Column)
	}
	if !global.App.AuditConfig.ENABLE_COLUMN_TIMESTAMP_TYPE && c.Tp == mysql.TypeTimestamp {
		return fmt.Errorf("列`%s`不允许定义TIMESTAMP类型[表`%s`]", c.Column, c.Table)
	}
	return nil
}

// 检查列not null
func (c *ColOptions) CheckColumnNotNull() error {
	if !global.App.AuditConfig.ENABLE_COLUMN_NOT_NULL {
		return nil
	}
	// 允许为NULL的类型
	allowNULLType := []byte{mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob, mysql.TypeJSON}
	// 是否允许时间类型设置为null
	if global.App.AuditConfig.ENABLE_COLUMN_TIME_NULL {
		allowNULLType = append(allowNULLType, []byte{mysql.TypeDatetime, mysql.TypeTimestamp, mysql.TypeDate, mysql.TypeYear}...)
	}
	// 列必须定义NOT NULL
	if !utils.IsByteContain(allowNULLType, c.Tp) && !c.NotNullFlag {
		return fmt.Errorf("列`%s`必须定义为`NOT NULL`[表`%s`]", c.Column, c.Table)
	}
	// 不合法的定义`NOT NULL DEFAULT NULL`
	if c.NotNullFlag && c.HasDefaultValue && c.DefaultIsNull {
		return fmt.Errorf("列`%s`不能定义`NOT NULL DEFAULT NULL`[表`%s`]", c.Column, c.Table)
	}
	return nil
}

// 检查列默认值
func (c *ColOptions) CheckColumnDefaultValue() error {
	// BLOB,TEXT,GEOMETRY,JSON类型不能设置默认值
	cannotSetDefaultValueType := []byte{mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob, mysql.TypeJSON, mysql.TypeGeometry}
	if utils.IsByteContain(cannotSetDefaultValueType, c.Tp) {
		if c.HasDefaultValue {
			return fmt.Errorf("列`%s`不能有一个默认值(BLOB/TEXT/GEOMETRY/JSON类型不能有一个默认值)[表`%s`]", c.Column, c.Table)
		}
	}
	// 列需要设置默认值
	if global.App.AuditConfig.CHECK_COLUMN_DEFAULT_VALUE && !c.HasDefaultValue && !utils.IsByteContain(cannotSetDefaultValueType, c.Tp) {
		return fmt.Errorf("列`%s`需要设置一个默认值[表`%s`]", c.Column, c.Table)
	}
	// 检查默认值(有默认值、且不为NULL)和数据类型是否匹配，Invalid default value
	if c.HasDefaultValue && !c.DefaultIsNull && !utils.IsByteContain(cannotSetDefaultValueType, c.Tp) {
		switch c.Tp {
		case mysql.TypeTiny, mysql.TypeShort, mysql.TypeInt24,
			mysql.TypeLong, mysql.TypeLonglong,
			mysql.TypeYear,
			mysql.TypeFloat, mysql.TypeDouble, mysql.TypeNewDecimal:
			// 验证string型默认值的合法性
			switch val := c.DefaultValue.(type) {
			case string:
				_, intErr := strconv.ParseInt(val, 10, 16)
				_, floatErr := strconv.ParseFloat(val, 64)
				if intErr != nil && floatErr != nil {
					return fmt.Errorf("列`%s`默认值和类型不匹配[表`%s`]", c.Column, c.Table)
				}
			}
		}
	}
	// 有默认值，配置了无效的默认值，如default current_timestamp
	if c.HasDefaultValue && !(c.Tp == mysql.TypeTimestamp || c.Tp == mysql.TypeDatetime) && c.DefaultValue == "current_timestamp" {
		return fmt.Errorf("列`%s`配置了无效的默认值(default current_timestamp)[表`%s`]", c.Column, c.Table)
	}
	return nil
}
