/*
@Time    :   2022/06/24 10:18:49
@Author  :   zongfei.fu
@Desc    :   获取目标数据库元信息
*/

package inspect

import (
	"errors"
	"fmt"
	"sqlSyntaxAudit/common/kv"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/controllers/parser"
	"strconv"
	"strings"

	mysqlapi "github.com/go-sql-driver/mysql"
)

// ShowCreateTable
func ShowCreateTable(table string, db *utils.DB, kv *kv.KVCache) (data interface{}, err error) {
	// 返回表结构
	data = kv.Get(table)
	if data != nil {
		return data, nil
	}
	query := fmt.Sprintf("show create table `%s`", table)
	result, err := db.FetchRows(query)
	if err != nil {
		return nil, err
	}
	var createStatement string
	for _, sql := range *result {
		// 表
		if _, ok := sql["Create Table"]; ok {
			createStatement = sql["Create Table"].(string)
		}
		// 视图
		if _, ok := sql["Create View"]; ok {
			createStatement = sql["Create View"].(string)
		}
	}

	var warns []error
	data, warns, err = parser.NewParse(createStatement, "", "")
	if len(warns) > 0 {
		return nil, fmt.Errorf("Parse Warning: %s", utils.ErrsJoin("; ", warns))
	}
	if err != nil {
		return nil, fmt.Errorf("SQL语法解析错误:%s", err.Error())
	}
	kv.Put(table, data)
	return data, nil
}

// descTable
func DescTable(table string, db *utils.DB) (error, string) {
	// 检查表是否存在，适用于确认当前实例当前库的表
	err := db.Exec(fmt.Sprintf("desc `%s`", table))
	if me, ok := err.(*mysqlapi.MySQLError); ok {
		if me.Number == 1146 {
			// 表不存在
			return err, fmt.Sprintf("表或视图`%s`不存在", table)
		} else if me.Number == 1045 {
			return err, fmt.Sprintf("访问目标数据库%s:%d失败,%s", db.Host, db.Port, err.Error())
		}
	}
	return nil, fmt.Sprintf("表或视图`%s`已经存在", table)
}

// verifyTable
func VerifyTable(table string, db *utils.DB) (error, string) {
	// 通过information_schema.tables检查表是否存在，适用于确认当前实例跨库的表
	result, err := db.FetchRows(fmt.Sprintf("select count(*) as count from information_schema.tables where table_name='%s'", table))
	if err != nil {
		return err, fmt.Sprintf("执行SQL失败,主机:%s:%d,错误:%s", db.Host, db.Port, err.Error())
	}
	var count int
	for _, row := range *result {
		count, _ = strconv.Atoi(row["count"].(string))
		break
	}
	if count == 0 {
		// 表不存在
		return errors.New("error"), fmt.Sprintf("表或视图`%s`不存在", table)
	}
	// 表存在
	return nil, fmt.Sprintf("表或视图`%s`已经存在", table)
}

// 获取DB变量
func GetDBVars(db *utils.DB) (map[string]string, error) {
	result, err := db.FetchRows("show variables where Variable_name in  ('innodb_large_prefix','version','character_set_database')")
	if err != nil {
		return nil, err
	}
	data := make(map[string]string)
	for _, row := range *result {
		if row["Variable_name"] == "version" {
			data["dbVersion"] = row["Value"].(string)
		}
		if row["Variable_name"] == "character_set_database" {
			data["dbCharset"] = row["Value"].(string)
		}
		if row["Variable_name"] == "innodb_large_prefix" {
			var largePrefix string
			switch row["Value"].(string) {
			case "0":
				largePrefix = "OFF"
			case "1":
				largePrefix = "ON"
			default:
				largePrefix = strings.ToUpper(row["Value"].(string))
			}
			data["largePrefix"] = largePrefix
		}
	}
	return data, nil
}
