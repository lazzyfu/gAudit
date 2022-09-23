- [表Options检查](#表options检查)
  - [检查项](#检查项)
- [主键检查](#主键检查)
  - [Case](#case)
  - [检查项](#检查项-1)
- [外键检查](#外键检查)
  - [检查项](#检查项-2)
- [审计字段检查](#审计字段检查)
  - [Case](#case-1)
  - [检查项](#检查项-3)
- [列Options检查](#列options检查)
  - [检查项](#检查项-4)
- [索引检查](#索引检查)
  - [检查项](#检查项-5)
- [DML检查](#dml检查)
  - [检查项](#检查项-6)
### 表Options检查
#### 检查项
| 关键字                               | 检查项                                        |
| :----------------------------------- | :-------------------------------------------- |
| N/A                                  | 检查表是否存在                                |
| enable_create_table_as               | 是否允许使用允许使用create table as语法       |
| enable_create_table_like             | 是否允许使用create table like语法             |
| ENABLE_CREATE_VIEW                   | 是否允许创建视图                              |
| MAX_TABLE_NAME_LENGTH                | 检查表名的长度                                |
| CHECK_IDENTIFIER                     | [a-zA-Z0-9_] 是否启用表名合法性检查           |
| CHECK_IDENTIFER_KEYWORD              | 对象名是否可以使用关键字                      |
| CHECK_TABLE_ENGINE                   | 是否启用存储引擎检查                          |
| TABLE_SUPPORT_ENGINE                 | 支持定义有效的存储引擎，如InnoDB/MyISAM/NDB等 |
| ENABLE_PARTITION_TABLE               | 是否支持分区表                                |
| CHECK_TABLE_COMMENT                  | 检查表是否有注释                              |
| TABLE_COMMENT_LENGTH                 | 检查表的注释长度，防止溢出                    |
| CHECK_TABLE_CHARSET                  | 是否启用字符集和排序规则检查                  |
| TABLE_SUPPORT_CHARSET                | 是否支持分区表                                |
| CHECK_TABLE_AUTOINCREMENT_INIT_VALUE | 检查建表时自增列初始值是否为1                 |
| TABLE_AT_LEAST_ONE_COLUMN            | 表至少要有一列，语法默认支持                  |
| ENABLE_DROP_COLS                     | 是否允许DROP列                                |
| N/A                                  | 检查drop的列是否存在                          |
| ENABLE_DROP_PRIMARYKEY               | 是否允许DROP主键                              |
| ENABLE_COLUMN_TYPE_CHANGE            | 是否允许变更数据类型                          |
| ENABLE_COLUMN_CHANGE_COLUMN_NAME                 | 是否禁止CHANGE操作                            |
| ENABLE_COLUMN_TYPE_CHANGE_COMPATIBLE                 | 是否开启change列类型兼容模式                            |
| ENABLE_INDEX_RENAME                  | 是否允许RENAME INDEX操作                      |
| ENABLE_RENAME_TABLE_NAME             | 是否允许RENAME表名                            |
| ENABLE_DROP_TABLE                    | 是否允许DROP表                                |
| ENABLE_TRUNCATE_TABLE                | 是否允许TRUNCATE表                            |



### 主键检查
#### Case
```sql
I_ID bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID'
```

#### 检查项
| 关键字                              | 检查项                   |
| :---------------------------------- | :----------------------- |
| CHECK_TABLE_PRIMARY_KEY             | 表是否有主键             |
| CHECK_TABLE_PRIMARY_KEY             | 表只能定义一个主键       |
| CHECK_PRIMARYKEY_USE_BIGINT         | 主键必须为bigint         |
| CHECK_PRIMARYKEY_USE_UNSIGNED       | 主键bigint必须为unsigned |
| CHECK_PRIMARYKEY_USE_AUTO_INCREMENT | 主键必须定义为自增       |
| N/A                                 | 主键必须定义NOT NULL     |

### 外键检查
#### 检查项
| 关键字             | 备注         |
| :----------------- | :----------- |
| ENABLE_FOREIGN_KEY | 是否启用外键 |

### 审计字段检查
#### Case
```sql
CREATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
UPDATED_AT datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
```

####  检查项
| 关键字                         | 备注                                                                 |
| :----------------------------- | :------------------------------------------------------------------- |
| CHECK_TABLE_AUDIT_TYPE_COLUMNS | 是否启用审计类型的字段检查，不检查字段名，仅检测字段定义是否符合要求 |


### 列Options检查
#### 检查项
| 关键字                       | 备注                                                                                                                                            |
| :--------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------- |
| MAX_COLUMN_NAME_LENGTH       | 列名的长度                                                                                                                                      |
| CHECK_IDENTIFIER             | [a-zA-Z0-9_] 是否启用列名合法性检查                                                                                                             |
| CHECK_IDENTIFER_KEYWORD      | 对象名是否可以使用关键字                                                                                                                        |
| CHECK_COLUMN_COMMENT         | 列是否有注释                                                                                                                                    |
| COLUMN_MAX_CHAR_LENGTH       | char长度大于N的时候需要改为varchar                                                                                                              |
| MAX_VARCHAR_LENGTH           | 最大允许定义的varchar长度                                                                                                                       |
| CHECK_COLUMN_FLOAT_DOUBLE    | 将float/double建议转成int/bigint/decimal等                                                                                                      |
| ENABLE_COLUMN_BLOB_TYPE      | 是否允许列的类型为BLOB/TEXT                                                                                                                     |
| ENABLE_COLUMN_JSON_TYPE      | 是否允许列的类型为JSON                                                                                                                          |
| ENABLE_COLUMN_TIMESTAMP_TYPE | 是否允许列的类型为TIMESTAMP                                                                                                                     |
| ENABLE_COLUMN_NOT_NULL       | 列是否定义为NOT NULL                                                                                                                            |
| N/A                          | TEXT/BLOB/JSON类型允许为NULL                                                                                                                    |
| ENABLE_COLUMN_TIME_NULL      | datetime/timestamp是否允许为NULL                                                                                                                |
| CHECK_COLUMN_DEFAULT_VALUE   | 列必须要有默认值                                                                                                                                |
| N/A                          | 不能定义`NOT NULL DEFAULT NULL`                                                                                                                 |
| N/A                          | BLOB,TEXT,GEOMETRY,JSON类型不能设置默认值                                                                                                       |
| N/A                          | 检查默认值(有默认值、且不为NULL)和数据类型是否匹配;默认值是否有效，如不为datetime和timestamp类型的字段配置了`default current_timestamp`         |
| CHECK_COLUMN_CHARSET         | 是否启用列的字符集检查;<br>列必须同时指定字符集和排序规则;<br>列的排序规则的前缀必须为字符集名;<br>mysql仅支持对char/varchar/enum/set指定字符集 |
| N/A                          | 列重复定义检查                                                                                                                                  |

### 索引检查
#### 检查项
| 关键字                       | 备注                                                                |
| :--------------------------- | :------------------------------------------------------------------ |
| CHECK_IDENTIFIER             | a-zA-Z0-9_] 是否启用索引名合法性检查                                |
| N/A                          | 索引名不能为空                                                      |
| CHECK_UNIQ_INDEX_PREFIX      | 是否启用唯一索引前缀检查                                            |
| UNQI_INDEX_PREFIX            | 唯一索引前缀检查                                                    |
| CHECK_SECONDARY_INDEX_PREFIX | 是否启用二级索引前缀检查                                            |
| SECONDARY_INDEX_PREFIX       | 二级索引前缀检查                                                    |
| CHECK_FULLTEXT_INDEX_PREFIX  | 是否启用全文索引前缀检查                                            |
| FULLTEXT_INDEX_PREFIX        | 全文索引前缀检查                                                    |
| MAX_INDEX_KEYS               | 检查二级索引的数量，包括唯一索引                                    |
| PRIMARYKEY_MAX_KEY_PARTS     | 检查主键数量                                                        |
| N/A                          | 检查是否重复定义了索引                                              |
| N/A                          | 创建索引时，指定的列必须存在                                        |
| N/A                          | 创建索引时，索引中的列不能重复                                      |
| N/A                          | 创建索引时，索引名不能重复                                          |
| N/A                          | 不能有重复的索引,即索引名不同,字段相同；冗余索引,如(a),(a,b)        |
| N/A                          | 查找重复的索引,即索引名不一样,但是定义的列一样,不区分大小写         |
| N/A                          | 查找冗余的索引,即索引名不一样,但是定义的列冗余,不区分大小写         |
| N/A                          | BLOB/TEXT类型不能设置为索引                                         |
| N/A                          | IndexLargePrefix检查，自适应DB版本和参数设置（innodb-large-prefix） |
| ENABLE_DROP_INDEXES          | 是否允许DROP索引                                                    |
| N/A                          | 检查drop的索引是否存在                                              |


### DML检查
#### 检查项
| 关键字                     | 备注                                       |
| :------------------------- | :----------------------------------------- |
| DISABLE_INSERT_INTO_SELECT | 是否禁止使用insert/replace into select语法 |
| DISABLE_ON_DUPLICATE       | 是否禁止使用insert into on duplicate语法   |
| DML_MUST_HAVE_WHERE        | DML语句必须要有where条件                   |
| DISABLE_REPLACE            | 是否允许使用replace语句                    |
| N/A                        | insert/replace语句必须指定列名             |
| MAX_INSERT_ROWS            | INSERT语句单次最多允许插入的行数           |
| DML_DISABLE_LIMIT          | delete/update语句是否能有LIMIT子句         |
| DML_DISABLE_ORDERBY        | delete/update语句是否能有orderby子句       |
| DML_DISABLE_SUBQUERY       | delete/update语句是否能有子查询            |
| CHECK_DML_JOIN_WITH_ON     | delete/update语句的JOIN操作是否要有ON条件  |
| MAX_AFFECTED_ROWS          | DML语句执行计划最大影响行数                |