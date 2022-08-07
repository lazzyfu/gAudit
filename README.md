# gAduit
gAudit是一个纯粹的SQL语法审核工具，支持MySQL/TiDB，通过解析SQL语法树实现自定义规则的审核规则。

#### 文档
- [快速开始](docs/start.md)
- [审核参数](docs/parameters.md)
- [审核规则](docs/rules.md)

### 语法解析器
* [tidb parser](https://github.com/pingcap/tidb/tree/master/parser)


#### 使用
```
curl --location --request POST '127.0.0.1:8081/api/v1/audit' \
--header 'Content-Type: application/json' \
--data-raw '{
    "db_user": "gaudit_rw",
    "db_password": "1234.com",
    "db_host": "127.0.0.1",
    "db_port": 3306,
    "db": "dbms_monitor",
    "timeout": 3000,
    "custom_audit_parameters": {"MAX_VARCHAR_LENGTH": 2000},
    "sqltext": "alter table slamonitor modify `address` varchar(16554) NOT NULL DEFAULT '\'''\'' COMMENT '\''主机'\''"
}
'
```
* db_user: 审核数据库用户
* db_password: 审核数据库密码
* db_host: 审核数据库地址
* db_port: 审核数据库端口
* db: 审核db
* timeout: 访问审核数据库超时时间，单位ms
* custom_audit_parameters: 自定义传递的审核参数，优先级: `自定义传递参数` > `template/config.json` > `config/config.go`
* sqltext: SQL语句，支持多条SQL语句，每条SQL语句以分号`;`分割，传递的语句越多，审核耗时越长，不建议一次超过1千条


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