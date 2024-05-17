/*
@Time    :   2022/06/23 16:16:32
@Author  :   xff
@Desc    :   表单
*/

package forms

type SyntaxAuditForm struct {
	DbUser            string                 `json:"db_user"`
	DbPassword        string                 `json:"db_password"`
	DbHost            string                 `json:"db_host"`
	DbPort            int                    `json:"db_port"`
	DB                string                 `json:"db"`
	Timeout           int64                  `json:"timeout"`                 // 连接数据库超时，单位ms
	CustomAuditParams map[string]interface{} `json:"custom_audit_parameters"` // 自定义的参数
	SqlText           string                 `json:"sqltext"`                 // 审计的SQL文本
	RequestID         string                 `json:"request_id"`              // 每次请求的ID
}

type ExtractTablesForm struct {
	SqlText   string `json:"sqltext"`    // SQL文本
	RequestID string `json:"request_id"` // 每次请求的ID
}
