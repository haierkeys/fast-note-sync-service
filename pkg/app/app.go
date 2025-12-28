package app

import (
	"reflect"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Ctx *gin.Context
}

type Pager struct {
	// 页码
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"pageSize"`
	// 总行数
	TotalRows int `json:"totalRows"`
}
type ListRes struct {
	// 数据清单
	List interface{} `json:"list"`
	// 翻页信息
	Pager Pager `json:"pager"`
}

type ResResult struct {
	// 业务状态码
	Code int `json:"code"`
	// 状态
	Status bool `json:"status"`
	// 失败&&成功消息
	Msg interface{} `json:"message"`
	// 数据集合
	Data interface{} `json:"data"`
}

type ResDetailsResult struct {
	// 业务状态码
	Code int `json:"code"`
	// 状态
	Status bool `json:"status"`
	// 失败&&成功消息
	Msg interface{} `json:"message"`
	// 错误格式数据
	Data interface{} `json:"data"`
	// 错误支付
	Details interface{} `json:"details"`
}

type ResVaultResult struct {
	// 业务状态码
	Code int `json:"code"`
	// 状态
	Status bool `json:"status"`
	// 失败&&成功消息
	Msg interface{} `json:"message"`
	// 数据集合
	Data  interface{} `json:"data"`
	Vault string      `json:"vault"`
}

type ResVaultDetailsResult struct {
	// 业务状态码
	Code int `json:"code"`
	// 状态
	Status bool `json:"status"`
	// 失败&&成功消息
	Msg interface{} `json:"message"`
	// 错误格式数据
	Data interface{} `json:"data"`
	// 错误支付
	Details interface{} `json:"details"`
	Vault   string      `json:"vault"`
}

type ResListResult struct {
	// 业务状态码
	Code int `json:"code"`
	// 状态
	Status bool `json:"status"`
	// 失败&&成功消息
	Msg interface{} `json:"message"`
	// 数据集合
	Data ListRes `json:"data"`
}

func NewResponse(ctx *gin.Context) *Response {
	return &Response{
		Ctx: ctx,
	}
}

// RequestParamStrParse 解析
func RequestParamStrParse(c *gin.Context, param any) {
	tParam := reflect.TypeOf(param).Elem()
	vParam := reflect.ValueOf(param).Elem()
	for i := 0; i < tParam.NumField(); i++ {
		name := tParam.Field(i).Name
		if nameType, ok := tParam.FieldByName(name); ok {
			dstName := nameType.Tag.Get("request")
			if dstName != "" {
				paramName := nameType.Tag.Get("form")
				if value, ok := c.GetQuery(paramName); ok {
					vParam.FieldByName(dstName).SetString(value)
				}
			}
		}
	}
}

// GetRequestIP 获取ip
func GetRequestIP(c *gin.Context) string {
	reqIP := c.ClientIP()
	if reqIP == "::1" {
		reqIP = "127.0.0.1"
	}
	return reqIP
}

func GetAccessHost(c *gin.Context) string {
	AccessProto := ""
	if proto := c.Request.Header.Get("X-Forwarded-Proto"); proto == "" {
		AccessProto = "http" + "://"
	} else {
		AccessProto = proto + "://"
	}
	return AccessProto + c.Request.Host
}

// ToResponse 输出到浏览器
func (r *Response) ToResponse(code *code.Code) {

	r.Ctx.Set("status_code", code.StatusCode())
	if code.HaveDetails() {
		details := strings.Join(code.Details(), ",")
		r.SendResultResponse(code.StatusCode(), ResDetailsResult{
			Code:    code.Code(),
			Status:  code.Status(),
			Msg:     code.Lang.GetMessage(),
			Data:    code.Data(),
			Details: details,
		})
	} else {
		r.SendResultResponse(code.StatusCode(), ResResult{
			Code:   code.Code(),
			Status: code.Status(),
			Msg:    code.Lang.GetMessage(),
			Data:   code.Data(),
		})
	}
}

func (r *Response) ToResponseList(code *code.Code, list interface{}, totalRows int) {

	r.Ctx.Set("status_code", code.StatusCode())

	r.SendResultResponse(code.StatusCode(), ResListResult{
		Code:   code.Code(),
		Status: code.Status(),
		Msg:    code.Lang.GetMessage(),
		Data: ListRes{
			List: list,
			Pager: Pager{
				Page:      GetPage(r.Ctx),
				PageSize:  GetPageSize(r.Ctx),
				TotalRows: totalRows,
			},
		},
	})
}

func (r *Response) SendResultResponse(statusCode int, content interface{}) {
	r.Ctx.JSON(statusCode, content)
}
