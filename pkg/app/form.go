package app

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/haierkeys/fast-note-sync-service/pkg/json"
)

type ValidError struct {
	Key     string
	Message string
}

type ValidErrors []*ValidError

func (v *ValidError) Error() string {
	return v.Message
}

func (v *ValidError) Field() string {
	return v.Key
}

func (v *ValidError) Map() map[string]string {
	return map[string]string{v.Key: v.Message}
}

func (v ValidErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func (v ValidErrors) Errors() []string {
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return errs
}

func (v ValidErrors) ErrorsToString() string {
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return strings.Join(errs, ",")
}

func (v ValidErrors) Maps() []map[string]string {
	var maps []map[string]string
	for _, err := range v {
		maps = append(maps, err.Map())
	}

	return maps
}

func (v ValidErrors) MapsToString() string {
	maps := v.Maps()
	re, _ := json.Marshal(maps)
	return string(re)
}

// BindAndValid bind request parameters and perform validation, supporting multiple languages
// BindAndValid 绑定请求参数并进行验证，支持多语言
func BindAndValid(c *gin.Context, obj interface{}) (bool, ValidErrors) {
	var errs ValidErrors

	// Use global validator for validation
	// 使用全局验证器进行验证
	if err := c.ShouldBind(obj); err != nil {
		// If verification fails, check error type
		// 如果验证失败，检查错误类型
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Get translator
			// 获取翻译器
			v := c.Value("trans")
			trans := v.(ut.Translator)

			// Iterate through validation errors and translate them
			// 遍历验证错误并进行翻译
			for _, validationErr := range validationErrors {
				translatedMsg := validationErr.Translate(trans) // 翻译错误消息
				errs = append(errs, &ValidError{
					Key:     validationErr.Field(),
					Message: translatedMsg,
				})
			}
		}

		return false, errs // Return validation error // 返回验证错误
	}

	return true, nil // Binding and validation both succeeded, returns true // 绑定和验证都成功，返回 true
}

// RequestParam extracts the specified parameter from the request without consuming or breaking the request body stream.
// RequestParam 从请求中提取指定参数（支持 Query、Form、JSON Body），且不消费或破坏请求体流。
func RequestParam(c *gin.Context, key string) string {
	if v := c.Query(key); v != "" {
		return v
	}
	if v := c.PostForm(key); v != "" {
		return v
	}
	if c.Request != nil && strings.Contains(c.ContentType(), "application/json") && c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			var bodyObj map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &bodyObj); err == nil {
				if val, ok := bodyObj[key]; ok {
					if strVal, ok := val.(string); ok {
						return strVal
					}
					switch valType := val.(type) {
					case float64:
						return strconv.FormatFloat(valType, 'f', -1, 64)
					case bool:
						return strconv.FormatBool(valType)
					default:
						return fmt.Sprintf("%v", val)
					}
				}
			}
		}
	}
	return ""
}
