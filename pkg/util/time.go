package util

import (
	"time"
)

// GetFirstDateOfMonth 获取传入的时间所在月份的第一天，即某月第一天的0点
// d: 传入的时间
// 返回值: 该月第一天的0点时间
func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTime(d)
}

// GetLastDateOfMonth 获取传入的时间所在月份的最后一天，即某月最后一天的0点
// d: 传入的时间
// 返回值: 该月最后一天的0点时间
func GetLastDateOfMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 1, -1)
}

// GetZeroTime 获取某一天的0点时间
// d: 传入的时间
// 返回值: 当天的0点时间
func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// GetEndTime 获取某一天的23:59:59时间
// d: 传入的时间
// 返回值: 当天的23:59:59时间
func GetEndTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, d.Location())
}

// GetLastDateOfNextMonth 获取传入时间的下个月最后一天
// d: 传入的时间
// 返回值: 下个月最后一天的时间
func GetLastDateOfNextMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 2, -1)
}

// Wait 等待指定的秒数
// num: 等待的秒数
func Wait(num float32) {
	tmpTime := time.Duration(num * 1000000000)
	time.Sleep(tmpTime)
}

// TimeParse 时间日期格式化
// layout: 时间格式
// in: 要解析的时间字符串
// 返回值: 解析后的时间对象
func TimeParse(layout string, in string) time.Time {
	local, _ := time.LoadLocation("Local")
	timer, _ := time.ParseInLocation(layout, in, local)
	return timer
}
