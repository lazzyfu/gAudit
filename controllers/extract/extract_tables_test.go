package extract

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/forms"
	"sqlSyntaxAudit/global"
	logger "sqlSyntaxAudit/middleware/log"
	"testing"
)

func init() {
	// 初始化配置
	global.App.AuditConfig = &config.AuditConfiguration{
		LogFilePath: "../../logs",
	}
	logger.Setup()
}

func TestChecker_Extract(t *testing.T) {
	tests := []struct {
		name    string
		form    forms.ExtractTablesForm
		wantErr error
		wantRes []ReturnData
	}{
		{
			name: "简单查询",
			form: forms.ExtractTablesForm{
				SqlText:   "select * from t1",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{"t1"},
					Type:   "SELECT",
					Query:  "select * from t1",
				},
			},
		},
		{
			name: "TIDB不支持的hint(mysql hint)",
			form: forms.ExtractTablesForm{
				SqlText:   "SELECT /*+ NO_RANGE_OPTIMIZATION(t3 PRIMARY, f2_idx) */ f1 FROM t3 WHERE f1 > 30 AND f1 < 33;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantErr: errors.New("Parse Warning: [parser:8061]Optimizer hint NO_RANGE_OPTIMIZATION is not supported by TiDB and is ignored"),
		},
		{
			name: "TIDB支持的hint",
			form: forms.ExtractTablesForm{
				SqlText:   "SELECT /*!40001 SQL_NO_CACHE */ * FROM `film`;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					// WITH子查询别名不应当放入Tables中
					Tables: []string{
						"film",
					},
					Type:  "SELECT",
					Query: "SELECT /*!40001 SQL_NO_CACHE */ * FROM `film`;",
				},
			},
		},
		{
			name: "With语句",
			form: forms.ExtractTablesForm{
				SqlText:   "WITH xm_gl AS ( SELECT * FROM products WHERE pname IN ( '小米电视机', '格力空调' ) ) SELECT avg( price ) FROM xm_gl;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					// WITH子查询别名不应当放入Tables中
					Tables: []string{
						"products",
					},
					Type:  "SELECT",
					Query: "WITH xm_gl AS ( SELECT * FROM products WHERE pname IN ( '小米电视机', '格力空调' ) ) SELECT avg( price ) FROM xm_gl;",
				},
			},
		},
		{
			name: "With语句别名与表名相同",
			form: forms.ExtractTablesForm{
				SqlText:   "WITH products AS ( SELECT * FROM products WHERE pname IN ( select name FROM `order` where user_id=1 ) ) SELECT avg( price ) FROM products;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					// WITH子查询别名不应当放入Tables中
					Tables: []string{
						"products",
						"order",
					},
					Type:  "SELECT",
					Query: "WITH products AS ( SELECT * FROM products WHERE pname IN ( select name FROM `order` where user_id=1 ) ) SELECT avg( price ) FROM products;",
				},
			},
		},
		{
			name: "SELECT未查询任何表",
			form: forms.ExtractTablesForm{
				SqlText:   "select 'hello';",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{},
					Type:   "SELECT",
					Query:  "select 'hello';",
				},
			},
		},
		{
			name: "简单JOIN",
			form: forms.ExtractTablesForm{
				SqlText:   "select t.table_schema,t.table_name,engine from information_schema.tables t inner join information_schema.columns c on t.table_schema=c.table_schema and t.table_name=c.table_name group by t.table_schema,t.table_name;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{
						"tables",
						"columns",
					},
					Type:  "SELECT",
					Query: "select t.table_schema,t.table_name,engine from information_schema.tables t inner join information_schema.columns c on t.table_schema=c.table_schema and t.table_name=c.table_name group by t.table_schema,t.table_name;",
				},
			},
		},
		{
			name: "标量子查询",
			form: forms.ExtractTablesForm{
				SqlText:   "select (select max(salary) from b where b.id=a.id) from a;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{
						"a",
						"b",
					},
					Type:  "SELECT",
					Query: "select (select max(salary) from b where b.id=a.id) from a;",
				},
			},
		},
		{
			name: "简单DELETE",
			form: forms.ExtractTablesForm{
				SqlText:   "delete from t1 where D_TIME >='2022-08-17 00:00:00' and D_TIME < '2022-08-18 00:00:00';",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{
						"t1",
					},
					Type:  "DELETE",
					Query: "delete from t1 where D_TIME >='2022-08-17 00:00:00' and D_TIME < '2022-08-18 00:00:00';",
				},
			},
		},
		{
			name: "关联DELETE",
			form: forms.ExtractTablesForm{
				SqlText:   "DELETE t1 FROM t1, t2 WHERE t1.id=t2.id",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{
						"t1",
						"t2",
					},
					Type:  "DELETE",
					Query: "DELETE t1 FROM t1, t2 WHERE t1.id=t2.id",
				},
			},
		},
		{
			name: "INSERT INTO SELECT",
			form: forms.ExtractTablesForm{
				SqlText:   "INSERT INTO T1 SELECT * FROM T2 WHERE id in (SELECT ID FROM T3)",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{
						"T1",
						"T2",
						"T3",
					},
					Type:  "INSERT",
					Query: "INSERT INTO T1 SELECT * FROM T2 WHERE id in (SELECT ID FROM T3)",
				},
			},
		},
		{
			name: "RENAME TABLE",
			form: forms.ExtractTablesForm{
				SqlText:   "RENAME TABLE t1 TO t2;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{
						"t1",
						"t2",
					},
					Type:  "RENAME TABLE",
					Query: "RENAME TABLE t1 TO t2;",
				},
			},
		},
		{
			name: "RENAME TABLE",
			form: forms.ExtractTablesForm{
				SqlText:   "CREATE VIEW v1 AS SELECT * FROM t1 WHERE c1 > 2;",
				RequestID: "78c25a06-3b34-4ecb-b9dd-7197078873c7",
			},
			wantRes: []ReturnData{
				{
					Tables: []string{
						"v1",
						"t1",
					},
					Type:  "CREATE VIEW",
					Query: "CREATE VIEW v1 AS SELECT * FROM t1 WHERE c1 > 2;",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Checker{
				Form: tt.form,
			}
			err, res := c.Extract(tt.form.RequestID)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				// 预期会有错误返回，就不需要进一步校验其它两个返回值了
				return
			}
			assert.Equal(t, tt.wantRes, res)
		})
	}
}
