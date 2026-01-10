package convert

import (
	"strconv"
	"strings"
)

type StrTo string

func (s StrTo) String() string {
	return string(s)
}

func (s StrTo) Int() (int, error) {
	v, err := strconv.Atoi(s.String())
	return v, err
}

func (s StrTo) MustInt() int {
	v, _ := s.Int()
	return v
}

func (s StrTo) UInt32() (uint32, error) {
	v, err := strconv.Atoi(s.String())
	return uint32(v), err
}
func (s StrTo) Int64() (int64, error) {
	v, err := strconv.Atoi(s.String())
	return int64(v), err
}

func (s StrTo) MustInt64() int64 {
	v, _ := s.Int64()
	return v
}

func (s StrTo) MustUInt32() uint32 {
	v, _ := s.UInt32()
	return v
}

// ToSize 将字符串转换为字节大小，支持 KB, MB, B 后缀
func (s StrTo) ToSize() (int64, error) {
	sizeStr := strings.ToUpper(strings.TrimSpace(s.String()))
	if sizeStr == "" {
		return 0, nil
	}

	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "B") {
		multiplier = 1
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	size, err := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)
	if err != nil {
		return 0, err
	}

	return size * multiplier, nil
}

// MustToSize 将字符串转换为字节大小，如果出错返回默认值
func (s StrTo) MustToSize(defaultVal int64) int64 {
	v, err := s.ToSize()
	if err != nil || v <= 0 {
		return defaultVal
	}
	return v
}
