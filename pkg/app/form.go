package app

import (
	"strings"

	"github.com/haierkeys/fast-note-sync-service/pkg/json"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
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
