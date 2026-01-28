package api_router

import (
	"expvar"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Expvar exports system runtime metrics
// Expvar 导出系统运行时指标
// Function name: Expvar
// 函数名: Expvar
// usage: Handles HTTP requests for system runtime metrics (expvar). Writes JSON data exported by expvar to the response.
// 函数使用说明: 处理获取系统运行时指标 (expvar) 的 HTTP 请求。将 expvar 导出的 JSON 数据写入响应。
// Parameters:
//   - c *gin.Context: Gin Context
//
// 参数说明:
//   - c *gin.Context: Gin 上下文
//
// Return:
//   - JSON: JSON data containing system metrics
//
// 返回值说明:
//   - JSON: 包含系统指标的 JSON 数据
func Expvar(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	first := true
	report := func(key string, value interface{}) {
		if !first {
			fmt.Fprintf(c.Writer, ",\n")
		}
		first = false
		if str, ok := value.(string); ok {
			fmt.Fprintf(c.Writer, "%q: %q", key, str)
		} else {
			fmt.Fprintf(c.Writer, "%q: %v", key, value)
		}
	}

	fmt.Fprintf(c.Writer, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		report(kv.Key, kv.Value)
	})
	fmt.Fprintf(c.Writer, "\n}\n")
}
