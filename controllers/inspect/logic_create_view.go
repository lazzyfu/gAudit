/*
@Time    :   2022/07/06 10:12:14
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/global"
)

// LogicCreateViewIsExist
func LogicCreateViewIsExist(v *TraverseCreateViewIsExist, r *Rule) {
	if !global.App.AuditConfig.ENABLE_CREATE_VIEW {
		r.Summary = append(r.Summary, fmt.Sprintf("不允许创建视图`%s`", v.View))
		r.IsSkipNextStep = true
		return
	}
	// 检查视图是否存在,如果视图存在,skip下面的检查
	if err, msg := DescTable(v.View, r.DB); err == nil {
		r.Summary = append(r.Summary, msg)
		r.IsSkipNextStep = true
	}
}
