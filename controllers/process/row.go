package process

import (
	"fmt"
	"gAudit/pkg/kv"

	"github.com/jinzhu/copier"
)

// RowSizeTooLarge
type PartSpecification struct {
	Column  string
	Tp      byte
	Elems   []string // Elems is the element list for enum and set type.
	Flen    int      // 字段长度
	Decimal int      // decimal字段专用,decimal(12,2)中的2
	Charset string   // 列字符集
}
type InnoDBRowSize struct {
	Table    string // 表名
	Engine   string // 表引擎
	Charset  string // 表字符集
	ColsMaps []PartSpecification
}

// https://dev.mysql.com/doc/refman/8.3/en/innodb-row-format.html
func (l *InnoDBRowSize) Check(kv *kv.KVCache) error {
	if l.Engine != "InnoDB" {
		return nil
	}

	// MySQL 表的内部具有65,535字节的最大行大小限制
	maxRowSize := 65535

	// version
	versionIns := DbVersion{kv.Get("dbVersion").(string)}

	// 计算列长度
	var maxSumRowsLength int

	// 判断字符集，当列字符集为空，使用表的字符集
	for _, i := range l.ColsMaps {
		// &{{riskcontrol_derived_variable_conf1 utf8mb4 [{i_id 3 [] 11 -1 } {ch_code 15 [] 200 -1 }]}}
		// 处理字符集为空的情况
		if len(i.Charset) == 0 {
			i.Charset = l.Charset
		}

		var instDataBytes DataBytes
		err := copier.CopyWithOption(&instDataBytes, i, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err != nil {
			return err
		}
		maxSumRowsLength += instDataBytes.Get(versionIns.Int())
	}
	// 判断是否触发了行大小限制
	if maxSumRowsLength > maxRowSize {
		return fmt.Errorf("表`%s`触发了Row Size Limit，最大行大小为%d，当前为%d（表存储引擎为%s）", l.Table, maxRowSize, maxSumRowsLength, l.Engine)
	}

	return nil
}
