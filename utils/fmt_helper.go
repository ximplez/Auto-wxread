package utils

import "time"

// FormatWeekdayCN 将时间转换为中文周格式
// 示例：周日、周一...周六
func FormatWeekdayCN(t time.Time) string {
	weekdays := [...]string{
		"周日",
		"周一",
		"周二",
		"周三",
		"周四",
		"周五",
		"周六",
	}
	return weekdays[t.Weekday()]
}
