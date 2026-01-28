package code

import (
	"fmt"
	"net/http"
)

// Code is an immutable error code object // Code 是一个不可变的错误码对象
// All With* methods return new instances, original object is not modified
// 所有 With* 方法都返回新实例，不修改原对象
type Code struct {
	code        int         // Status code // 状态码
	status      bool        // Status // 状态
	Lang        lang        // Error message // 错误消息
	msg         string      // Error message // 错误消息
	data        interface{} // Data // 数据
	vault       string
	haveVault   bool     // Whether it contains Vault // 是否含有Vault
	haveData    bool     // Whether it contains Data // 是否含有Data
	details     []string // Error detail information // 错误详细信息
	haveDetails bool     // Whether it contains details // 是否含有详情
	context     string
	haveContext bool // Whether it contains Context // 是否含有Context
}

var codes = map[int]string{}
var maxcode = 0

func NewError(code int, reset ...bool) *Code {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("错误码 %d 已经存在，请更换一个", code))
	}

	l := getLang(code)
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

func NewSuss(code int) *Code {
	if _, ok := sussCodes[code]; ok {
		panic(fmt.Sprintf("成功码 %d 已经存在，请更换一个", code))
	}
	l := getLang(code)
	sussCodes[code] = l.GetMessage()
	if code > maxcode {
		maxcode = code
	}

	return &Code{code: code, status: true, Lang: l}
}

func getLang(code int) lang {
	return lang{
		zh_cn: zh_cn_messages[code],
		en:    en_messages[code],
	}
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

// WithData returns a new Code instance containing specified data
// WithData 返回一个包含指定数据的新 Code 实例
// Original object will not be modified (immutable design)
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

// WithVault returns a new Code instance containing specified vault
// WithVault 返回一个包含指定 vault 的新 Code 实例
// Original object will not be modified (immutable design)
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

// WithDetails returns a new Code instance containing specified details
// WithDetails 返回一个包含指定详情的新 Code 实例
// Original object will not be modified (immutable design)
// 原对象不会被修改（不可变设计）
func (e *Code) WithDetails(details ...string) *Code {
	// Create a copy of details to avoid shared underlying array
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

// WithContext returns a new Code instance containing specified context
// WithContext 返回一个包含指定上下文的新 Code 实例
// Original object will not be modified (immutable design)
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
