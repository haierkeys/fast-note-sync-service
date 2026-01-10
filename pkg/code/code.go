package code

import (
	"fmt"
	"net/http"
)

// Code 是一个不可变的错误码对象
// 所有 With* 方法都返回新实例，不修改原对象
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

func (e *Code) Error() string {
	return e.Msg()
}

func (e *Code) ErrorWithErr(err ...error) string {
	if len(err) > 0 {
		return e.Msg() + ": " + err[0].Error()
	}
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

// WithData 返回一个包含指定数据的新 Code 实例
// 原对象不会被修改（不可变设计）
func (e *Code) WithData(data interface{}) *Code {
	return &Code{
		code:        e.code,
		status:      e.status,
		Lang:        e.Lang,
		msg:         e.msg,
		data:        data,
		haveData:    true,
		vault:       e.vault,
		haveVault:   e.haveVault,
		details:     e.details,
		haveDetails: e.haveDetails,
		context:     e.context,
		haveContext: e.haveContext,
	}
}

// WithVault 返回一个包含指定 vault 的新 Code 实例
// 原对象不会被修改（不可变设计）
func (e *Code) WithVault(vault string) *Code {
	return &Code{
		code:        e.code,
		status:      e.status,
		Lang:        e.Lang,
		msg:         e.msg,
		data:        e.data,
		haveData:    e.haveData,
		vault:       vault,
		haveVault:   true,
		details:     e.details,
		haveDetails: e.haveDetails,
		context:     e.context,
		haveContext: e.haveContext,
	}
}

// WithDetails 返回一个包含指定详情的新 Code 实例
// 原对象不会被修改（不可变设计）
func (e *Code) WithDetails(details ...string) *Code {
	// 创建 details 的副本，避免共享底层数组
	newDetails := make([]string, len(details))
	copy(newDetails, details)

	return &Code{
		code:        e.code,
		status:      e.status,
		Lang:        e.Lang,
		msg:         e.msg,
		data:        e.data,
		haveData:    e.haveData,
		vault:       e.vault,
		haveVault:   e.haveVault,
		details:     newDetails,
		haveDetails: true,
		context:     e.context,
		haveContext: e.haveContext,
	}
}

// WithContext 返回一个包含指定上下文的新 Code 实例
// 原对象不会被修改（不可变设计）
func (e *Code) WithContext(context string) *Code {
	return &Code{
		code:        e.code,
		status:      e.status,
		Lang:        e.Lang,
		msg:         e.msg,
		data:        e.data,
		haveData:    e.haveData,
		vault:       e.vault,
		haveVault:   e.haveVault,
		details:     e.details,
		haveDetails: e.haveDetails,
		context:     context,
		haveContext: true,
	}
}

func (e *Code) StatusCode() int {
	return http.StatusOK
}
