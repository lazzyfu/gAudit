- [最佳实践](#最佳实践)
  - [自定义NOT NULL](#自定义not-null)
    - [参数](#参数)
      - [ENABLE_COLUMN_NOT_NULL](#enable_column_not_null)
      - [ENABLE_COLUMN_TIME_NULL](#enable_column_time_null)
    - [允许为NULL的类型](#允许为null的类型)
  - [审计字段](#审计字段)
    - [参数](#参数-1)
      - [CHECK_TABLE_AUDIT_TYPE_COLUMNS](#check_table_audit_type_columns)
  - [限制指定的表进行DDL/DML语法审核](#限制指定的表进行ddldml语法审核)
## 最佳实践
### 自定义NOT NULL
#### 参数
##### ENABLE_COLUMN_NOT_NULL
启用字段not null检查，设置为true后，会要求字段设置为not null

例子:
```sql
address varchar(128) not null default '' comment '地址'
```

##### ENABLE_COLUMN_TIME_NULL
时间类型字段允许设置为null。例如一些业务字段需要设置为允许为null，比如部分业务的时间字段不希望实现`magic`

例子:
```sql
finish_at datetime default null comment '完成时间'
```

#### 允许为NULL的类型
* text
* blob
* json

例子
```sql
remark text comment '备注'
```

### 审计字段
建表时必须要有`创建时间`和`更新时间`, 该如何实现呢

#### 参数
##### CHECK_TABLE_AUDIT_TYPE_COLUMNS
启用审计类型的字段, 必须定义2个审计字段，要求
* DEFAULT CURRENT_TIMESTAMP
* DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

例子
> 字段名和注释名不做要求
```sql
`UPDATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
`CREATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
```

### 限制指定的表进行DDL/DML语法审核
一些特殊的场景，研发希望部分表不能提交DDL和DML工单，场景举例：

用户业务库：
  -  库：db_users
  -  表：tbl_users
  -  类型：主表
  -  业务读写该表

支付业务库：
  -  库：db_pay
  -  表：tbl_users
  -  类型：备表
  -  业务只读该表，数据和表结构从主表同步

上述场景下，我们可以配置db_pay.tbl_users禁止语法审核，可以保证对db_pay.tbl_users的DDL和DML操作无法提交，保证数据一致性。

```json
"DISABLE_AUDIT_DDL_TABLES": [
      {
        "DB": "db_pay",
        "Reason": "限制审核和提交,请联系支付业务研发",
        "Tables": [
          "tbl_users"
        ]
      },
    ],
"DISABLE_AUDIT_DML_TABLES": [
  {
    "DB": "db_pay",
    "Reason": "限制审核和提交,请联系支付业务研发",
    "Tables": [
      "tbl_users"
    ]
  }
],
```
