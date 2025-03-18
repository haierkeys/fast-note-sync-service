package code

import (
	"fmt"
	"net/http"
)

type Code struct {
	// 状态码
	code int
	// 状态
	status bool
	// 错误消息
	Lang lang
	// 错误消息
	msg string
	// 数据
	data interface{}
	// 错误详细信息
	details []string
	// 是否含有详情
	haveDetails bool
}

var codes = map[int]string{}
var maxcode = 0

func NewError(code int, l lang) *Code {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("错误码 %d 已经存在，请更换一个", code))
	}

	codes[code] = l.GetMessage()

	if code > maxcode {
		maxcode = code
	}

	return &Code{code: code, status: false, Lang: l}
}

func incr(code int) int {
	if code > maxcode {
		return code
	} else {
		return maxcode + 1
	}
}

var sussCodes = map[int]string{}

func NewSuss(code int, l lang) *Code {
	if _, ok := sussCodes[code]; ok {
		panic(fmt.Sprintf("成功码 %d 已经存在，请更换一个", code))
	}
	sussCodes[code] = l.GetMessage()
	if code > maxcode {
		maxcode = code
	}

	return &Code{code: code, status: true, Lang: l}
}

func (e *Code) Error() string {
	return e.Msg()
}

func (e *Code) Code() int {
	return e.code
}

func (e *Code) Status() bool {
	return e.status
}

func (e *Code) Msg() string {
	return e.Lang.GetMessage()
}

func (e *Code) Msgf(args []interface{}) string {
	return fmt.Sprintf(e.msg, args...)
}

func (e *Code) Details() []string {
	return e.details
}

func (e *Code) Data() interface{} {
	return e.data
}

func (e *Code) HaveDetails() bool {
	return e.haveDetails
}

func (e *Code) WithData(data interface{}) *Code {
	e.data = data
	return e
}

func (e *Code) WithDetails(details ...string) *Code {
	e.haveDetails = true
	e.details = []string{}
	for _, d := range details {
		e.details = append(e.details, d)
	}
	return e
}

func (e *Code) StatusCode() int {
	return http.StatusOK
}
