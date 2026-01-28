package code

import (
	"errors"
	"fmt"
	"reflect"
)

// lang type, used to store English and Chinese text
// lang 类型，用来存储英文和中文文本
type lang struct {
	en    string // English // 英文
	zh_cn string // Chinese // 中文
}

// Default language is English // 默认语言为英文
var lng = "en"

const FALLBACK_LNG = "en"

// GetMessage method returns the corresponding message according to the passed language
// GetMessage 方法根据传入的语言返回相应的消息
func (l lang) GetMessage() string {
	if lng == "" {
		lng = FALLBACK_LNG
	}
	// Get language field
	// 获取语言字段
	val := reflect.ValueOf(l)
	field := val.FieldByName(lng)
	// If the language field is valid and not empty, return the message in that language
	// 如果语言字段有效且非空，返回该语言的消息
	if field.IsValid() && field.String() != "" {
		return field.String()
	}
	// If the specified language is invalid, return the message of the fallback language
	// 如果指定语言无效，返回回退语言的消息
	fallbackField := val.FieldByName(FALLBACK_LNG)
	if fallbackField.IsValid() && fallbackField.String() != "" {
		return fallbackField.String()
	}
	// If the fallback language has no message either, return the default error message
	// 如果回退语言也没有消息，返回默认的错误信息
	return fmt.Sprintf("No message available for language: %s", lng)
}

// GetSupportedLanguages function returns all languages supported by the lang type
// GetSupportedLanguages 函数返回 lang 类型支持的所有语言
func GetSupportedLanguages() []string {
	var languages []string
	// Get the fields of the lang type through reflection
	// 通过反射获取 lang 类型的字段
	typ := reflect.TypeOf(lang{})
	// Traverse the fields of the struct and get the field names
	// 遍历结构体的字段，获取字段名
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		languages = append(languages, field.Name)
	}
	return languages
}

// SetGlobalDefaultLang sets the global default language
// 设置全局默认语言
func SetGlobalDefaultLang(language string) error {
	// Supported language list
	// 支持的语言列表
	supportedLanguages := GetSupportedLanguages()

	// Check if the language is in the list of supported languages
	// 检查语言是否在支持的语言列表中
	isValidLang := false
	for _, lang := range supportedLanguages {
		if language == lang {
			isValidLang = true
			break
		}
	}
	// If the language is valid, set the global language
	// 如果语言有效，设置全局语言
	if isValidLang {
		lng = language
		return nil
	}
	// If the language is invalid, return an error and set it to the default language
	// 如果语言无效，返回错误并设置为默认语言
	lng = FALLBACK_LNG
	return errors.New("unsupported language type, set defaulting to " + FALLBACK_LNG)
}

// GetGlobalDefaultLang gets the global default language
// 设置全局默认语言
func GetGlobalDefaultLang() string {
	return lng
}
