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
	data  interface{}
	vault string
	// 是否含有Vault
	haveVault bool
	// 是否含有Data
	haveData bool
	// 错误详细信息
	details []string
	// 是否含有详情
	haveDetails bool
	context     string
	// 是否含有Context
	haveContext bool
}

var codes = map[int]string{}
var maxcode = 0

func CodeReset() {

}

func NewError(code int, l lang, reset ...bool) *Code {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("错误码 %d 已经存在，请更换一个", code))
	}

	codes[code] = l.GetMessage()

	if code > maxcode {
		maxcode = code
	}

	if len(reset) > 0 && reset[0] {
		maxcode = 0
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

func (e *Code) Reset() *Code {
	e.data = nil
	e.haveDetails = false
	e.haveData = false
	e.details = []string{}
	e.haveVault = false
	e.vault = ""
	e.haveContext = false
	e.context = ""
	return e
}

// Clone 创建一个新的 Code 副本
func (e *Code) Clone() *Code {
	// 创建一个新的副本,而不是修改原对象
	return &Code{
		code:   e.code,
		status: e.status,
		Lang:   e.Lang,
		msg:    e.msg,
		// 其他字段保持默认零值
		data:        nil,
		vault:       "",
		haveVault:   false,
		haveData:    false,
		details:     []string{},
		haveDetails: false,
		context:     "",
		haveContext: false,
	}
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

func (e *Code) Vault() string {
	return e.vault
}

func (e *Code) Context() string {
	return e.context
}

func (e *Code) HaveDetails() bool {
	return e.haveDetails
}

func (e *Code) HaveData() bool {
	return e.haveData
}

func (e *Code) HaveVault() bool {
	return e.haveVault
}

func (e *Code) HaveContext() bool {
	return e.haveContext
}

func (e *Code) WithData(data interface{}) *Code {
	e.haveData = true
	e.data = data
	return e
}

func (e *Code) WithVault(vault string) *Code {
	e.haveVault = true
	e.vault = vault
	return e
}

func (e *Code) WithDetails(details ...string) *Code {
	e.haveDetails = true
	e.details = []string{}

	e.details = append(e.details, details...)

	return e
}

func (e *Code) WithContext(context string) *Code {
	e.haveContext = true
	e.context = context
	return e
}

func (e *Code) StatusCode() int {
	return http.StatusOK
}
