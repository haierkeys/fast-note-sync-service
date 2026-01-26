package app

import (
	"reflect"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// VersionInfo 版本信息
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

// BaseRes 是统一的响应结构：Code/Status/Msg/Data
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

// RequestParamStrParse 解析
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

// ToResponse 输出到浏览器：统一使用 BaseRes，根据情况设置 Details 与 Vault
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
		// 假设 codeObj.Vault() 返回可序列化的值（string 或 struct 等）
		content.Vault = codeObj.Vault()
	}

	r.send(codeObj.StatusCode(), content)
}

// ToResponseList 输出列表响应，使用 ListRes 作为 Data；同样支持 Vault 动态添加
func (r *Response) ToResponseList(codeObj *code.Code, list interface{}, totalRows int) {
	r.Ctx.Set("status_code", codeObj.StatusCode())

	content := Res{
		Code:    codeObj.Code(),
		Status:  codeObj.Status(),
		Message: codeObj.Lang.GetMessage(),
		Data: ListRes{
			List: list,
			Pager: Pager{
				Page:      GetPage(r.Ctx),
				PageSize:  GetPageSize(r.Ctx),
				TotalRows: totalRows,
			},
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
