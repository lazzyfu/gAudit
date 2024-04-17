
# gAudit
![GO](https://img.shields.io/badge/go-1.21.8-brightgreen.svg?style=flat-square)
![Download](https://img.shields.io/github/downloads/lazzyfu/gAudit/total?style=flat-square)
![License](https://img.shields.io/github/license/lazzyfu/gAudit?style=flat-square)
[![release](https://img.shields.io/github/v/release/lazzyfu/gAudit.svg)](https://github.com/lazzyfu/gAudit/releases)
<img alt="Github Stars" src="https://img.shields.io/github/stars/lazzyfu/gAudit?logo=github">

gAudit是基于golang语言实现的一个SQL语法审核工具，支持MySQL/TiDB，通过解析SQL语法树实现语法规则审核。

## 文档
- [快速开始](docs/start.md)
- [审核参数](docs/parameters.md)
- [审核规则](docs/rules.md)
- [最佳实践](docs/practice.md)

## 语法解析器
* [tidb parser](https://github.com/pingcap/tidb/tree/master/parser)


## 使用方法
> 服务端口依赖于您启动指定的端口，下面8081端口为举例

| API                                         | 请求方法 | 用途     | 备注                                     |
| ------------------------------------------- | -------- | -------- | ---------------------------------------- |
| http://127.0.0.1:8081/api/v1/audit          | POST     | 语法审核 | 支持DDL/DML语句，支持一次提交多条SQL语句 |
| http://127.0.0.1:8081/api/v1/extract-tables | POST     | 提取表名 | 支持DDL/DML语句，支持一次提交多条SQL语句 |

### 语法审核
#### POST请求
```bash
curl --request POST '127.0.0.1:8081/api/v1/audit' \
--header 'Content-Type: application/json' \
--data '{
    "db_user": "gaudit_rw",
    "db_password": "1234.com",
    "db_host": "127.0.0.1",
    "db_port": 3306,
    "db": "dbms_monitor",
    "timeout": 3000,
    "custom_audit_parameters": {"MAX_VARCHAR_LENGTH": 2000},
    "sqltext": "alter table slamonitor modify `address` varchar(16554) NOT NULL DEFAULT '\'''\'' COMMENT '\''主机'\''"
}
' | jq
```
* db_user: 审核数据库用户
* db_password: 审核数据库密码
* db_host: 审核数据库地址
* db_port: 审核数据库端口
* db: 审核db
* timeout: 访问审核数据库超时时间，单位ms
* custom_audit_parameters: 自定义传递的审核参数，优先级: `自定义传递参数` > `template/config.json` > `config/config.go`
* sqltext: SQL语句，支持多条SQL语句，每条SQL语句以分号`;`分割


#### 输出
```json
{
    "request_id": "0a2392e4-ee3f-4f9c-9da1-3906ae4521c9",
    "code": "0000",
    "data": [
        {
            "summary": [
                "列`host`最大允许定义的varchar长度为2000,当前varchar长度为16554[表`slamonitor`]"
            ],
            "level": "WARN",
            "affected_rows": 0,
            "type": "DDL",
            "finger_id": "4B3E7A0DCAE81036",
            "query": "alter table slamonitor modify `host` varchar(16554) NOT NULL DEFAULT '' COMMENT '主机'"
        }
    ],
    "message": "success"
}
```

### 提取表名
> 支持DML/DDL、union以及更复杂的查询等

#### POST请求
```bash
curl --location --request POST '127.0.0.1:8081/api/v1/extract-tables' \
--header 'Content-Type: application/json' \
--data '{
    "sqltext": "alter table t1 add name varchar(100);select * from (select id,name from tt1 join tt2 on tt1.id=tt2.id where tt1.id > 100) as xx;UPDATE product p, product_price pp SET pp.price = p.price * 0.8 WHERE p.productid= pp.productId;"
}' | jq .
```

#### 输出
```json
{
  "request_id": "cb9e5249-c77c-4320-bbfb-9fe0a9391da7",
  "code": "0000",
  "data": [
    {
      "tables": [
        "t1"
      ],
      "type": "ALTER TABLE",
      "query": "alter table t1 add name varchar(100);"
    },
    {
      "tables": [
        "tt1",
        "tt2"
      ],
      "type": "SELECT",
      "query": "select * from (select id,name from tt1 join tt2 on tt1.id=tt2.id where tt1.id > 100) as xx;"
    },
    {
      "tables": [
        "product",
        "product_price"
      ],
      "type": "UPDATE",
      "query": "UPDATE product p, product_price pp SET pp.price = p.price * 0.8 WHERE p.productid= pp.productId;"
    }
  ],
  "message": "success"
}
```

### Python调用接口模板
> 请根据自己的需求进行封装改造即可

文件位置`template/python_api.py`


## 致谢
- [PingCAP](https://github.com/pingcap/tidb/tree/master/parser)
