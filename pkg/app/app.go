package app

import (
	"reflect"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// VersionInfo version information // 版本信息
type VersionInfo struct {
	Version   string `json:"version"`
	GitTag    string `json:"gitTag"`
	BuildTime string `json:"buildTime"`
}

type CheckVersionInfo struct {
	VersionIsNew         bool   `json:"versionIsNew"`
	VersionNewName       string `json:"versionNewName"`
	VersionNewLink       string `json:"versionNewLink"`
	PluginVersionIsNew   bool   `json:"pluginVersionIsNew"`
	PluginVersionNewName string `json:"pluginVersionNewName"`
	PluginVersionNewLink string `json:"pluginVersionNewLink"`
}

type Response struct {
	Ctx *gin.Context
}

type Pager struct {
	Page      int `json:"page"`      // Page number // 页码
	PageSize  int `json:"pageSize"`  // Page size // 每页数量
	TotalRows int `json:"totalRows"` // Total rows // 总行数
}

// PaginationRequest pagination request parameters for Swagger // 分页请求参数（用于 Swagger）
type PaginationRequest struct {
	Page     int `json:"page" form:"page" query:"page"`             // Page number // 页码
	PageSize int `json:"pageSize" form:"pageSize" query:"pageSize"` // Page size // 每页数量
}

type ListRes struct {
	List  interface{} `json:"list"`  // Data list // 数据清单
	Pager Pager       `json:"pager"` // Pagination info // 翻页信息
}

// Res is the unified response structure: Code/Status/Msg/Data
// Optional fields Vault and Details use omitempty (will not be serialized if nil)
// Res 是统一的响应结构：Code/Status/Msg/Data
// 可选字段 Vault 与 Details 使用 omitempty（nil 则不会被序列化）
type Res struct {
	Code    int         `json:"code"`
	Status  bool        `json:"status"`
	Message interface{} `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Details interface{} `json:"details,omitempty"`
	Vault   interface{} `json:"vault,omitempty"`
	Context interface{} `json:"context,omitempty"`
}

func NewResponse(ctx *gin.Context) *Response {
	return &Response{
		Ctx: ctx,
	}
}

// RequestParamStrParse parses request parameters
// RequestParamStrParse 解析
// Keep original behavior
// 保持原有行为
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

// GetRequestIP gets the request IP
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

// ToResponse output to browser: unified use of Res, set Details and Vault as needed
// ToResponse 输出到浏览器：统一使用 Res，根据情况设置 Details 与 Vault
func (r *Response) ToResponse(codeObj *code.Code) {
	r.Ctx.Set("status_code", codeObj.StatusCode())

	content := Res{
		Code:    codeObj.Code(),
		Status:  codeObj.Status(),
		Message: codeObj.Lang.GetMessage(),
		Data:    codeObj.Data(),
	}

	if codeObj.HaveDetails() {
		content.Details = strings.Join(codeObj.Details(), ",")
	}

	if codeObj.HaveVault() {
		// Assume codeObj.Vault() returns a serializable value (string, struct, etc.)
		// 假设 codeObj.Vault() 返回可序列化的值（string 或 struct 等）
		content.Vault = codeObj.Vault()
	}

	r.send(codeObj.StatusCode(), content)
}

// ToResponseList outputs list response using ListRes as Data; also supports dynamic Vault addition
// ToResponseList 输出列表响应，使用 ListRes 作为 Data；同样支持 Vault 动态添加
func (r *Response) ToResponseList(codeObj *code.Code, list interface{}, totalRows int) {
	r.Ctx.Set("status_code", codeObj.StatusCode())

	content := Res{
		Code:    codeObj.Code(),
		Status:  codeObj.Status(),
		Message: codeObj.Lang.GetMessage(),
		Data: ListRes{
			List:  list,
			Pager: *NewPager(r.Ctx, totalRows),
		},
	}

	if codeObj.HaveVault() {
		content.Vault = codeObj.Vault()
	}

	r.send(codeObj.StatusCode(), content)
}

func (r *Response) send(statusCode int, content interface{}) {
	r.Ctx.JSON(statusCode, content)
}
