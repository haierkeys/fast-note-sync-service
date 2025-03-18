/**
  @author: haierkeys
  @since: 2022/9/14
  @desc: //TODO
**/

package gin_tools

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

func RequestParams(c *gin.Context) (map[string]interface{}, error) {

	const defaultMemory = 32 << 20
	contentType := c.ContentType()

	var (
		dataMap  = make(map[string]interface{})
		queryMap = make(map[string]interface{})
		postMap  = make(map[string]interface{})
	)

	// @see gin@v1.7.7/binding/query.go ==> func (queryBinding) Bind(req *httpclient.Request, obj interface{})
	for k := range c.Request.URL.Query() {
		queryMap[k] = c.Query(k)
	}

	if "application/json" == contentType {
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// @see gin@v1.7.7/binding/json.go ==> func (jsonBinding) Bind(req *httpclient.Request, obj interface{})
		if c.Request != nil && c.Request.Body != nil {
			if err := json.NewDecoder(c.Request.Body).Decode(&postMap); err != nil {
				return nil, err
			}
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else if "multipart/form-data" == contentType {
		// @see gin@v1.7.7/binding/form.go ==> func (formMultipartBinding) Bind(req *httpclient.Request, obj interface{})
		if err := c.Request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
		for k, v := range c.Request.PostForm {
			if len(v) > 1 {
				postMap[k] = v
			} else if len(v) == 1 {
				postMap[k] = v[0]
			}
		}
	} else {
		// ParseForm 解析 URL 中的查询字符串，并将解析结果更新到 r.Form 字段
		// 对于 POST 或 PUT 请求，ParseForm 还会将 body 当作表单解析，
		// 并将结果既更新到 r.PostForm 也更新到 r.Form。解析结果中，
		// POST 或 PUT 请求主体要优先于 URL 查询字符串（同名变量，主体的值在查询字符串的值前面）
		// @see gin@v1.7.7/binding/form.go ==> func (formBinding) Bind(req *httpclient.Request, obj interface{})
		if err := c.Request.ParseForm(); err != nil {
			return nil, err
		}
		if err := c.Request.ParseMultipartForm(defaultMemory); err != nil {
			if err != http.ErrNotMultipart {
				return nil, err
			}
		}
		for k, v := range c.Request.PostForm {
			if len(v) > 1 {
				postMap[k] = v
			} else if len(v) == 1 {
				postMap[k] = v[0]
			}
		}
	}

	var mu sync.RWMutex
	for k, v := range queryMap {
		mu.Lock()
		dataMap[k] = v
		mu.Unlock()
	}
	for k, v := range postMap {
		mu.Lock()
		dataMap[k] = v
		mu.Unlock()
	}

	return dataMap, nil
}
