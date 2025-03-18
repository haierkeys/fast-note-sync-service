package timex

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

type Time time.Time

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 2 {
		*t = Time(time.Time{})
		return
	}

	now, err := time.Parse(`"`+TimeFormat+`"`, string(data))
	*t = Time(now)
	return
}

func (t Time) MarshalJSON() ([]byte, error) {
	tTime := time.Time(t)
	// 如果时间值是空或者0值 返回为null 如果写空字符串会报出异常时间
	// 下面是修复0001-01-01问题的
	if &t == nil || t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", tTime.Format(TimeFormat))), nil

}

func (t *Time) IsZero() bool {
	return time.Time(*t).IsZero()
}

func (t Time) Value() (driver.Value, error) {
	if t.String() == "0000-00-00 00:00:00" {
		return nil, nil
	}
	if t.String() == "0001-01-01 00:00:00" {
		return nil, nil
	}
	return time.Time(t).Format(TimeFormat), nil
}

func (t *Time) Scan(v any) error {
	timeValue, ok := v.(time.Time)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal time value:", v))
	}
	*t = Time(timeValue)
	return nil

}

func (t Time) String() string {
	return time.Time(t).Format(TimeFormat)
}

func (t Time) StringSource() string {
	return time.Time(t).String()
}

func Now() Time {
	return Time(time.Now())
}

// After reports whether the time instant t is after u.
func (t Time) After(u Time) bool {
	ts := time.Time(t)
	return ts.After(time.Time(u))
}

// Before reports whether the time instant t is before u.
func (t Time) Before(u Time) bool {
	ts := time.Time(t)
	return ts.Before(time.Time(u))
}

// Equal reports whether t and u represent the same time instant.
// Two times can be equal even if they are in different locations.
// For example, 6:00 +0200 and 4:00 UTC are Equal.
// See the documentation on the Time type for the pitfalls of using == with
// Time values; most code should use Equal instead.
func (t Time) Equal(u Time) bool {
	ts := time.Time(t)
	return ts.Equal(time.Time(u))
}

// Add returns the time t+d.
func (t Time) Add(d time.Duration) Time {
	ts := time.Time(t)
	return Time(ts.Add(d))
}

func Since(t Time) time.Duration {
	ts := time.Time(t)
	return time.Since(ts)
}
