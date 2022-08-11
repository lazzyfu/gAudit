## Quick start
- [Quick start](#quick-start)
  - [不支持语法](#不支持语法)
  - [下载最新发行版](#下载最新发行版)
  - [创建配置文件](#创建配置文件)
  - [修改系统配置](#修改系统配置)
  - [启动服务](#启动服务)
  - [目标数据库创建审核账号](#目标数据库创建审核账号)
  - [自定义审核参数](#自定义审核参数)
  - [使用](#使用)
  - [输出](#输出)


### 不支持语法
- `GEOMETRY`数据类型

### 下载最新发行版
> 请移步到Releases下载最新版本

```bash
wget https://github.com/lazzyfu/gAudit/releases/download/v1.0.1/gAudit-linux-v1.0.1.tar.gz

mkdir /usr/local/gAudit
tar -jxf gAudit-linux-v1.0.1.tar.gz -C /usr/local/gAudit
```
### 创建配置文件
> 根据您自己的需求调整审核参数

模板文件: template/config.json
格式: json
审核参数: 请参考[审核参数](parameters.md)进行自定义增加或调整
文件：`/usr/local/gAudit/config.json`


### 修改系统配置
`vim /usr/local/gAudit/config.json`
```json
"ListenAddress": "127.0.0.1:80",
"LogFilePath": "./logs",
"LogLevel": "debug",
```

### 启动服务
> 您可以使用supervisor进行管理
```
chmod +x gAudit
./gAudit -config template/config.json &
```

### 目标数据库创建审核账号
> 请根据您的实际情况修改账号、主机和密码
> 每个需要审核的数据库均需要创建
```sql
create user 'gaudit_rw'@'%' identified by '1234.com';
GRANT SELECT, INSERT, UPDATE, DELETE ON *.* TO 'gaudit_rw'@'%'
```

### 自定义审核参数
> 有时需要临时放开某个审核规则，不希望每次修改配置文件然后去重启gAudit服务

**审核参数生效优先级**
`自定义传递参数（custom_audit_parameters）` > `template/config.json` > `config/config.go`

- **config/config.go**
系统内置审核参数，优先级最低，可以被`template/config.json`覆盖

- **template/config.json**
自定义的模板参数文件，您启动时加载的配置文件，可以被`自定义传递参数（custom_audit_parameters）`覆盖

- **custom_audit_parameters**
POST请求时自定义传参，优先级最高。支持一次传递多个审核参数（系统参数除外）

**custom_audit_parameters使用方法**
> 请参考[审核参数](parameters.md), 参数名不区分大小写
- 不允许使用关键字
- 必须要有审核字段
```
"custom_audit_parameters": {"check_identifer_keyword": true,"check_table_audit_type_columns": true},
```

### 使用
> 这里通过API的形式提交审核
```
curl --request POST '127.0.0.1:8081/api/v1/audit' \
--header 'Content-Type: application/json' \
--data '
{
    "db_user": "gaudit_rw",
    "db_password": "1234.com",
    "db_host": "127.0.0.1",
    "db_port": 3306,
    "db": "test",
    "timeout": 3000,
    "custom_audit_parameters": {"check_identifer_keyword": true,"check_table_audit_type_columns": true},
    "sqltext": "CREATE TABLE `meta_cluster` (\n  `id` tinyint(3) unsigned NOT NULL,\n  `env` varchar(32) NOT NULL DEFAULT '\'''\'' COMMENT '\''环境'\'',\n  `cluster_name` varchar(128) NOT NULL DEFAULT '\'''\'' COMMENT '\''集群名'\'',\n  `cluster_domain` varchar(128) NOT NULL DEFAULT '\'''\'' COMMENT '\''集群域名'\'',\n  `datacenter` varchar(128) NOT NULL DEFAULT '\'''\'' COMMENT '\''数据中心'\'',\n  `region` varchar(128) NOT NULL DEFAULT '\'''\'' COMMENT '\''区域'\'',\n  `hostname` varchar(128) NOT NULL DEFAULT '\'''\'' COMMENT '\''主机名'\'',\n  `promotion_rule` varchar(128) NOT NULL DEFAULT '\''prefer'\'' COMMENT '\''晋升规则，可选值: prefer/neutral/prefer_not/must_not'\'',\n  `port` int(11) NOT NULL DEFAULT '\''0'\'' COMMENT '\''端口'\'',\n  `D_UPDATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '\''更新时间'\'',\n  `D_CREATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '\''创建时间'\'',\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='\''orchestrator实例元数据表'\''"
}'
```
* db_user: 审核数据库用户
* db_password: 审核数据库密码
* db_host: 审核数据库地址
* db_port: 审核数据库端口
* db: 审核db
* timeout: 访问审核数据库超时时间，单位ms
* custom_audit_parameters: 自定义传递的审核参数，优先级: `自定义传递参数` > `template/config.json` > `config/config.go`
* sqltext: SQL语句，支持多条SQL语句，每条SQL语句以分号`;`分割，传递的语句越多，审核耗时越长，建议一次不要超过1千条


### 输出
```json
{
    "request_id": "652ed930-f06b-4dfe-a58a-e5f138afaeb2",
    "code": "0000",
    "data": [
        {
            "summary": [
                "表`meta_cluster`的自增列初始值必须显示指定且设置为1「例如:AUTO_INCREMENT=1」",
                "表`meta_cluster`的主键id必须定义unsigned",
                "表`meta_cluster`的主键`id`必须定义auto_increment",
                "表`meta_cluster`的主键id必须定义NOT NULL",
                "列`address`必须要有注释[表`meta_cluster`]",
                "列`address`需要设置一个默认值[表`meta_cluster`]",
                "列`datacenter`推荐设置为varchar(128)[表`meta_cluster`]",
                "列`promotion_rule`需要设置一个默认值[表`meta_cluster`]",
                "列`port`命名不允许使用关键字[表`meta_cluster`]",
                "表`meta_cluster`最多允许定义3个二级索引,当前二级索引个数为4",
                "表`meta_cluster`的索引`idx_adress`超出了innodb-large-prefix限制,当前索引长度为21317字节,最大限制为3072字节,当前字符集为utf8mb4「可使用前缀索引,如:Field(length)」"
            ],
            "level": "WARN",
            "affected_rows": 0,
            "type": "DDL",
            "finger_id": "5F286049EF7A887F",
            "query": "CREATE TABLE `meta_cluster` (\n  `id` bigint, address varchar(5200), `sex` enum('boy','girl') DEFAULT  NULL comment '性别',\n  `env` varchar(32) NOT NULL DEFAULT '' COMMENT '环境',\n  `cluster_name` varchar(128) NOT NULL DEFAULT '' COMMENT '集群名',\n  `cluster_domain` varchar(128) NOT NULL DEFAULT '' COMMENT '集群域名',\n  `datacenter` char(128) NOT NULL DEFAULT '' COMMENT '数据中心',\n  `region` varchar(128) NOT NULL DEFAULT '' COMMENT '区域',\n  `hostname` varchar(128) NOT NULL DEFAULT '' COMMENT '主机名',\n  `promotion_rule` varchar(128) NOT NULL COMMENT '晋升规则，可选值：prefer/neutral/prefer_not/must_not',\n  `port` int(11) NOT NULL DEFAULT '0' COMMENT '端口',\n  `D_UPDATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',\n  `D_CREATED_AT` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',\n  PRIMARY KEY (`id`), key idx_adress (address,sex,cluster_name), key `idx_name` (env, sex), key idx_sex (`sex`), KEY `idx_datacenter` (`datacenter`,cluster_domain(32))\n) ENGINE=InnoDB DEFAULT CHARACTER set utf8mb4 collate utf8mb4_general_ci COMMENT='orchestrator实例元数据表'"
        }
    ],
    "message": "success"
}
```