package checker

// 返回数据格式
type ReturnData struct {
	Summary      []string `json:"summary"` // 规则摘要
	Level        string   `json:"level"`   // 提醒级别,INFO/WARN/ERROR
	AffectedRows int      `json:"affected_rows"`
	Type         string   `json:"type"`
	FingerId     string   `json:"finger_id"`
	Query        string   `json:"query"` // 原始SQL
}
