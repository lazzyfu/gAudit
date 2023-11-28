# -*- coding:utf-8 -*-
# edit by fuzongfei
# python version >= 3.6
# 接口模板，请根据自己场景封装即可

from operator import itemgetter
import requests
import json

GAUDIT_API = "http://127.0.0.1:8082/api/v1/audit"


class GAuditApi(object):
    """gAudit语法检查接口"""

    def __init__(self, cfg=None, sql_text=None):
        # 目标数据库连接串配置
        self.cfg = cfg
        # 传入的SQL
        self.sql_text = sql_text
        self.cfg["sqltext"] = self.sql_text

    def request(self):
        # post
        header = {
            "Content-Type": "application/json"
        }

        try:
            resp = requests.post(
                GAUDIT_API,
                data=json.dumps(self.cfg, ensure_ascii=False).encode('utf-8'),
                headers=header
            )
        except requests.exceptions.ConnectionError as err:
            return 500, f"请求审核服务器gAudit异常，请联系DBA，错误信息:{err.args[0]}"

        if resp.status_code == 200:
            return resp.status_code, resp.json()

        return resp.status_code, f"请求审核服务器gAudit异常，请联系DBA;Code:{resp.status_code} Reason:{resp.reason}"

    def check(self):
        """判断语法检查是否通过
        返回值: status, data, msg
        """
        status_code, data = self.request()
        if status_code == 200:
            if data["code"] != "0000":
                return False, None, data["message"]
            keys = ['level']
            levels = [itemgetter(*keys)(row) for row in data["data"]]
            if all([i == "INFO" for i in levels]):
                return True, data["data"], None
            return False, data["data"], None
        return False, None, data


# 使用
sql_text = "delete from test_case where id > 1"
cfg = {
    "db_user": "sqlsyntaxaudit_rw",
    "db_password": "1234.com",
    "db_host": "127.0.0.1",
    "db_port": 3306,
    "db": "test",
    "timeout": 3000,
    "custom_audit_parameters": {
        "MAX_AFFECTED_ROWS": 1
    }
}

gaudit = GAuditApi(
    cfg=cfg,
    sql_text=sql_text
)
status, data, msg = gaudit.check()

print(f"检查是否通过: {status}")
print(f"返回数据: {json.dumps(data, indent=4, ensure_ascii=False)}")
print(f"msg: {msg}")

"""
 ~/Desktop/github/gAudit/template/ [fzf] python3 python_api.py
检查是否通过: False
返回数据: [
    {
        "summary": [
            "当前DELETE语句最大影响或扫描行数超过了最大允许值1【建议您将语句拆分为多条，保证每条语句影响或扫描行数小于最大允许值1】"
        ],
        "level": "WARN",
        "affected_rows": 2,
        "type": "DML",
        "finger_id": "D3A87C5D8BFAE066",
        "query": "delete from test_case where id > 1"
    }
]
msg: None
"""
