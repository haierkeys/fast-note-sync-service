package convert

import "strconv"

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
