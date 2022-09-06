- [审核参数](#审核参数)
  - [系统参数](#系统参数)
    - [ListenAddress](#listenaddress)
    - [LogFilePath](#logfilepath)
    - [LogLevel](#loglevel)
  - [审核参数](#审核参数-1)
    - [MAX_TABLE_NAME_LENGTH](#max_table_name_length)
    - [CHECK_TABLE_COMMENT](#check_table_comment)
    - [TABLE_COMMENT_LENGTH](#table_comment_length)
    - [CHECK_IDENTIFIER](#check_identifier)
    - [CHECK_IDENTIFER_KEYWORD](#check_identifer_keyword)
    - [CHECK_TABLE_CHARSET](#check_table_charset)
    - [TABLE_SUPPORT_CHARSET](#table_support_charset)
    - [CHECK_TABLE_ENGINE](#check_table_engine)
    - [TABLE_SUPPORT_ENGINE](#table_support_engine)
    - [ENABLE_PARTITION_TABLE](#enable_partition_table)
    - [CHECK_TABLE_PRIMARY_KEY](#check_table_primary_key)
    - [TABLE_AT_LEAST_ONE_COLUMN](#table_at_least_one_column)
    - [CHECK_TABLE_AUDIT_TYPE_COLUMNS](#check_table_audit_type_columns)
    - [enable_create_table_as](#enable_create_table_as)
    - [enable_create_table_like](#enable_create_table_like)
    - [ENABLE_FOREIGN_KEY](#enable_foreign_key)
    - [CHECK_TABLE_AUTOINCREMENT_INIT_VALUE](#check_table_autoincrement_init_value)
    - [ENABLE_CREATE_VIEW](#enable_create_view)
    - [MAX_COLUMN_NAME_LENGTH](#max_column_name_length)
    - [CHECK_COLUMN_CHARSET](#check_column_charset)
    - [CHECK_COLUMN_COMMENT](#check_column_comment)
    - [COLUMN_MAX_CHAR_LENGTH](#column_max_char_length)
    - [MAX_VARCHAR_LENGTH](#max_varchar_length)
    - [ENABLE_COLUMN_BLOB_TYPE](#enable_column_blob_type)
    - [ENABLE_COLUMN_JSON_TYPE](#enable_column_json_type)
    - [ENABLE_COLUMN_TIMESTAMP_TYPE](#enable_column_timestamp_type)
    - [CHECK_PRIMARYKEY_USE_BIGINT](#check_primarykey_use_bigint)
    - [CHECK_PRIMARYKEY_USE_UNSIGNED](#check_primarykey_use_unsigned)
    - [CHECK_PRIMARYKEY_USE_AUTO_INCREMENT](#check_primarykey_use_auto_increment)
    - [ENABLE_COLUMN_NOT_NULL](#enable_column_not_null)
    - [ENABLE_COLUMN_TIME_NULL](#enable_column_time_null)
    - [CHECK_COLUMN_DEFAULT_VALUE](#check_column_default_value)
    - [CHECK_COLUMN_FLOAT_DOUBLE](#check_column_float_double)
    - [ENABLE_COLUMN_TYPE_CHANGE](#enable_column_type_change)
    - [ENABLE_COLUMN_CHANGE](#enable_column_change)
    - [CHECK_UNIQ_INDEX_PREFIX](#check_uniq_index_prefix)
    - [CHECK_SECONDARY_INDEX_PREFIX](#check_secondary_index_prefix)
    - [CHECK_FULLTEXT_INDEX_PREFIX](#check_fulltext_index_prefix)
    - [UNQI_INDEX_PREFIX](#unqi_index_prefix)
    - [SECONDARY_INDEX_PREFIX](#secondary_index_prefix)
    - [FULLTEXT_INDEX_PREFIX](#fulltext_index_prefix)
    - [SECONDARY_INDEX_MAX_KEY_PARTS](#secondary_index_max_key_parts)
    - [PRIMARYKEY_MAX_KEY_PARTS](#primarykey_max_key_parts)
    - [MAX_INDEX_KEYS](#max_index_keys)
    - [ENABLE_INDEX_RENAME](#enable_index_rename)
    - [ENABLE_DROP_COLS](#enable_drop_cols)
    - [ENABLE_DROP_INDEXES](#enable_drop_indexes)
    - [ENABLE_DROP_PRIMARYKEY](#enable_drop_primarykey)
    - [ENABLE_DROP_TABLE](#enable_drop_table)
    - [ENABLE_TRUNCATE_TABLE](#enable_truncate_table)
    - [ENABLE_RENAME_TABLE_NAME](#enable_rename_table_name)
    - [ENABLE_MYSQL_MERGE_ALTER_TABLE](#enable_mysql_merge_alter_table)
    - [ENABLE_TIDB_MERGE_ALTER_TABLE](#enable_tidb_merge_alter_table)
    - [DML_MUST_HAVE_WHERE](#dml_must_have_where)
    - [DML_DISABLE_LIMIT](#dml_disable_limit)
    - [DML_DISABLE_ORDERBY](#dml_disable_orderby)
    - [DML_DISABLE_SUBQUERY](#dml_disable_subquery)
    - [CHECK_DML_JOIN_WITH_ON](#check_dml_join_with_on)
    - [EXPLAIN_RULE](#explain_rule)
    - [MAX_AFFECTED_ROWS](#max_affected_rows)
    - [MAX_INSERT_ROWS](#max_insert_rows)
    - [DISABLE_REPLACE](#disable_replace)
    - [DISABLE_INSERT_INTO_SELECT](#disable_insert_into_select)
    - [DISABLE_ON_DUPLICATE](#disable_on_duplicate)
    - [DISABLE_AUDIT_DML_TABLES](#disable_audit_dml_tables)
    - [DISABLE_AUDIT_DDL_TABLES](#disable_audit_ddl_tables)
## 审核参数
### 系统参数
> 不支持通过接口`custom_audit_parameters`传递的参数
#### ListenAddress 
描述: 服务侦听地址
默认值: 127.0.0.1:8081

#### LogFilePath
描述: 日志文件路径
默认值: ./logs

#### LogLevel
描述: 日志级别
默认值: info
可选值: debug/info/warn/error

### 审核参数
> [Identifier Length Limits](https://dev.mysql.com/doc/refman/8.0/en/identifier-length.html)
> 支持通过接口`custom_audit_parameters`传递的参数
#### MAX_TABLE_NAME_LENGTH
描述: 检查表名的长度
默认值: info
可选值: 1~64

#### CHECK_TABLE_COMMENT
描述: 检查表是否有注释
默认值: true
可选值: true/false

#### TABLE_COMMENT_LENGTH
描述: 表的注释长度
默认值: 64
可选值: 1~512

#### CHECK_IDENTIFIER
描述: 对象名必须使用字符串范围，匹配正则[a-zA-Z0-9_]
默认值: true
可选值: true/false

#### CHECK_IDENTIFER_KEYWORD
描述: 对象名是否可以使用关键字
默认值: false
可选值: true/false

#### CHECK_TABLE_CHARSET
描述: 是否检查表的字符集和排序规则
默认值: true
可选值: true/false

#### TABLE_SUPPORT_CHARSET
描述: 表支持的字符集
默认值: [{"charset": "utf8", "recommend": "utf8_general_ci"}, {"charset": "utf8mb4", "recommend": "utf8mb4_general_ci"}]
可选值: DB支持的字符集
CASE:
```sql
default character set utf8mb4 collate utf8mb4_general_ci
```

#### CHECK_TABLE_ENGINE
描述: 是否检查表的存储引擎
默认值: true
可选值: true/false

#### TABLE_SUPPORT_ENGINE
描述: 表支持的存储引擎
默认值: ["InnoDB"]
可选值: DB支持的存储引擎
CASE:
```sql
ENGINE=InnoDB
```

#### ENABLE_PARTITION_TABLE
描述: 是否启用分区表
默认值: false
可选值: true/false

#### CHECK_TABLE_PRIMARY_KEY
描述: 检查表是否有主键
默认值: true
可选值: true/false

#### TABLE_AT_LEAST_ONE_COLUMN
描述: 表至少要有一列，语法默认支持，调整参数无效
默认值: true
可选值: true/false

#### CHECK_TABLE_AUDIT_TYPE_COLUMNS
描述: 是否启用审计类型的字段检查，不检查字段名,仅检测字段定义是否符合要求，审计字段一般为`创建时间`/`更新时间`字段。
默认值: true
可选值: true/false
CASE：
```sql
CREATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
UPDATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
```

#### enable_create_table_as
描述: 是否允许create table as语法
默认值: false
可选值: true/false

#### enable_create_table_like
描述: 是否允许create table like语法
默认值: false
可选值: true/false

#### ENABLE_FOREIGN_KEY
描述: 是否启用外键
默认值: false
可选值: true/false

#### CHECK_TABLE_AUTOINCREMENT_INIT_VALUE
描述: 检查建表是自增列初始值是否为1
默认值: true
可选值: true/false

#### ENABLE_CREATE_VIEW
描述: 是否支持创建和使用视图
默认值: true
可选值: true/false

#### MAX_COLUMN_NAME_LENGTH
描述: 列名的长度
默认值: 32
可选值: 1~64

#### CHECK_COLUMN_CHARSET
描述: 是否检查列的字符集
默认值: true
可选值: true/false

#### CHECK_COLUMN_COMMENT
描述: 是否检查列的注释
默认值: true
可选值: true/false

#### COLUMN_MAX_CHAR_LENGTH
描述: char长度大于N的时候需要改为varchar
默认值: 64
可选值: 1~254

#### MAX_VARCHAR_LENGTH
描述: 最大允许定义的varchar长度
默认值: 65534
可选值: 1~65534

#### ENABLE_COLUMN_BLOB_TYPE
描述: 是否允许列的类型为BLOB/TEXT
默认值: true
可选值: true/false

#### ENABLE_COLUMN_JSON_TYPE
描述: 是否允许列的类型为JSON
默认值: true
可选值: true/false

#### ENABLE_COLUMN_TIMESTAMP_TYPE
描述: 是否允许列的类型为TIMESTAMP
默认值: false
可选值: true/false

#### CHECK_PRIMARYKEY_USE_BIGINT
描述: 主键是否为bigint
默认值: true
可选值: true/false

#### CHECK_PRIMARYKEY_USE_UNSIGNED
描述: 主键bigint是否为unsigned
默认值: true
可选值: true/false

#### CHECK_PRIMARYKEY_USE_AUTO_INCREMENT
描述: 主键是否定义为自增
默认值: true
可选值: true/false

#### ENABLE_COLUMN_NOT_NULL
描述: 列是否定义为NOT NULL
默认值: true
可选值: true/false

#### ENABLE_COLUMN_TIME_NULL
描述: 是否允许时间类型定义为NULL
默认值: true
可选值: true/false

#### CHECK_COLUMN_DEFAULT_VALUE
描述: 列必须要有默认值
默认值: true
可选值: true/false

#### CHECK_COLUMN_FLOAT_DOUBLE
描述: 将float/double转成int/bigint/decimal等
默认值: true
可选值: true/false

#### ENABLE_COLUMN_TYPE_CHANGE
描述: 是否允许变更列类型
默认值: false
可选值: true/false

#### ENABLE_COLUMN_CHANGE
描述: 是否允许CHANGE操作
默认值: false
可选值: true/false

#### CHECK_UNIQ_INDEX_PREFIX
描述: 是否检查唯一索引前缀,如唯一索引必须以uniq_为前缀
默认值: true
可选值: true/false

#### CHECK_SECONDARY_INDEX_PREFIX
描述: 是否检查二级索引前缀,如普通索引必须以idx_为前缀
默认值: true
可选值: true/false

#### CHECK_FULLTEXT_INDEX_PREFIX
描述: 是否检查全文索引前缀,如全文索引必须以full_为前缀
默认值: true
可选值: true/false

#### UNQI_INDEX_PREFIX
描述: 定义唯一索引前缀，不区分大小写
默认值: UNIQ_

#### SECONDARY_INDEX_PREFIX
描述: 定义二级索引前缀，不区分大小写
默认值: IDX_

#### FULLTEXT_INDEX_PREFIX
描述: 定义全文索引前缀，不区分大小写
默认值: FULL_

#### SECONDARY_INDEX_MAX_KEY_PARTS
描述: 组成二级索引的列数不能超过指定的个数,包括唯一索引
默认值: 8

#### PRIMARYKEY_MAX_KEY_PARTS
描述: 组成主键索引的列数不能超过指定的个数
默认值: 1

#### MAX_INDEX_KEYS
描述: 最多有N个索引,包括唯一索引/二级索引
默认值: 12

#### ENABLE_INDEX_RENAME
描述: 是否允许rename索引名
默认值: false
可选值: true/false

#### ENABLE_DROP_COLS
描述: 是否允许DROP列
默认值: true
可选值: true/false

#### ENABLE_DROP_INDEXES
描述: 是否允许DROP索引
默认值: true
可选值: true/false

#### ENABLE_DROP_PRIMARYKEY
描述: 是否允许DROP主键
默认值: false
可选值: true/false

#### ENABLE_DROP_TABLE
描述: 是否允许DROP TABLE
默认值: true
可选值: true/false

#### ENABLE_TRUNCATE_TABLE
描述: 是否允许TRUNCATE TABLE
默认值: true
可选值: true/false

#### ENABLE_RENAME_TABLE_NAME
描述: 是否允许rename表名
默认值: false
可选值: true/false

#### ENABLE_MYSQL_MERGE_ALTER_TABLE
描述: MySQL同一个表的多个ALTER是否合并为单条语句
默认值: true
可选值: true/false

#### ENABLE_TIDB_MERGE_ALTER_TABLE
描述: TiDB同一个表的多个ALTER是否合并为单条语句
默认值: false
可选值: true/false

#### DML_MUST_HAVE_WHERE
描述: DML语句必须有where条件
默认值: true
可选值: true/false

#### DML_DISABLE_LIMIT
描述: DML语句中不允许有LIMIT
默认值: true
可选值: true/false

#### DML_DISABLE_ORDERBY
描述: DML语句中不允许有orderby
默认值: true
可选值: true/false

#### DML_DISABLE_SUBQUERY
描述: DML语句不能有子查询
默认值: true
可选值: true/false

#### CHECK_DML_JOIN_WITH_ON
描述: DML的JOIN语句必须有ON语句
默认值: true
可选值: true/false

#### EXPLAIN_RULE
描述: explain判断受影响行数时使用的规则("first", "max")。 "first": 使用第一行的explain结果作为受影响行数, "max": 使用explain结果中的最大值作为受影响行数
默认值: first
可选值: first/max

#### MAX_AFFECTED_ROWS
描述: 最大影响行数，默认100
默认值: 100

#### MAX_INSERT_ROWS
描述: 一次最多允许insert的行, eg: insert into tbl(col,...) values(row1), (row2)...
默认值: 100

#### DISABLE_REPLACE
描述: 是否禁用replace语句
默认值: true
可选值: true/false

#### DISABLE_INSERT_INTO_SELECT
描述: 是否禁用insert/replace into select语法
默认值: true
可选值: true/false

#### DISABLE_ON_DUPLICATE
描述: 是否禁止insert on duplicate语法
默认值: true
可选值: true/false

#### DISABLE_AUDIT_DML_TABLES
适用场景: 多个库直接数据同步的主备表
描述: 禁止对指定的表进行DML审核
默认值: 空
配置例子:
```json
"DISABLE_AUDIT_DML_TABLES": [
    {
        "DB": "test",
        "Tables": [
            "t1",
            "t2"
        ],
        "Reason": "xxx业务研发禁止审核和提交,请联系xxx"
    }
]
```

#### DISABLE_AUDIT_DDL_TABLES
适用场景: 多个库直接数据同步的主备表
描述: 禁止对指定的表进行DDL审核
默认值: 空
配置例子:
```json
"DISABLE_AUDIT_DDL_TABLES": [
    {
        "DB": "test",
        "Tables": [
            "t1"
        ],
        "Reason": "xxx业务研发禁止审核和提交,请联系xxx"
    },
    {
        "DB": "test1",
        "Tables": [
            "c1",
            "c2"
        ],
        "Reason": "xxx业务研发禁止审核和提交,请联系xxx"
    }
]
```

