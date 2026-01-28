package util

import (
	"strconv"
	"strings"
	"time"
)

// GetFirstDateOfMonth gets the first day of the month for the given time, which is 0:00 on the first day of the month
// GetFirstDateOfMonth 获取传入的时间所在月份的第一天，即某月第一天的0点
// d: given time
// d: 传入的时间
// return: 0:00 on the first day of that month
// 返回值: 该月第一天的0点时间
func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTime(d)
}

// GetLastDateOfMonth gets the last day of the month for the given time, which is 0:00 on the last day of the month
// GetLastDateOfMonth 获取传入的时间所在月份的最后一天，即某月最后一天的0点
// d: given time
// d: 传入的时间
// return: 0:00 on the last day of that month
// 返回值: 该月最后一天的0点时间
func GetLastDateOfMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 1, -1)
}

// GetZeroTime gets 0:00 time of a certain day
// GetZeroTime 获取某一天的0点时间
// d: given time
// d: 传入的时间
// return: 0:00 time of that day
// 返回值: 当天的0点时间
func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// GetEndTime gets 23:59:59 time of a certain day
// GetEndTime 获取某一天的23:59:59时间
// d: given time
// d: 传入的时间
// return: 23:59:59 time of that day
// 返回值: 当天的23:59:59时间
func GetEndTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, d.Location())
}

// GetLastDateOfNextMonth gets the last day of the next month for the given time
// GetLastDateOfNextMonth 获取传入时间的下个月最后一天
// d: given time
// d: 传入的时间
// return: last day of the next month
// 返回值: 下个月最后一天的时间
func GetLastDateOfNextMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 2, -1)
}

// Wait waits for specified number of seconds
// Wait 等待指定的秒数
// num: number of seconds to wait
// num: 等待的秒数
func Wait(num float32) {
	tmpTime := time.Duration(num * 1000000000)
	time.Sleep(tmpTime)
}

// TimeParse time and date formatting
// TimeParse 时间日期格式化
// layout: time format
// layout: 时间格式
// in: time string to be parsed
// in: 要解析的时间字符串
// return: parsed time object
// 返回值: 解析后的时间对象
func TimeParse(layout string, in string) time.Time {
	local, _ := time.LoadLocation("Local")
	timer, _ := time.ParseInLocation(layout, in, local)
	return timer
}

// ParseDuration parses duration string, supports 'd' (day) suffix
// ParseDuration 解析时间字符串，支持 'd' (天) 后缀
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	// If it is pure numbers, default to seconds
	// 如果是纯数字，默认为秒
	if _, err := strconv.Atoi(s); err == nil {
		s += "s"
	}
	return time.ParseDuration(s)
}
