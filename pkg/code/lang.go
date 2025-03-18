package code

import (
	"errors"
	"fmt"
	"reflect"
)

// lang 类型，用来存储英文和中文文本
type lang struct {
	en string // 英文
	zh string // 中文
}

// 默认语言为英文
var lng = "zh"

const FALLBACK_LNG = "zh"

// getMessage 方法根据传入的语言返回相应的消息
func (l lang) GetMessage() string {
	if lng == "" {
		lng = FALLBACK_LNG
	}
	// 获取语言字段
	val := reflect.ValueOf(l)
	field := val.FieldByName(lng)
	// 如果语言字段有效且非空，返回该语言的消息
	if field.IsValid() && field.String() != "" {
		return field.String()
	}
	// 如果指定语言无效，返回回退语言的消息
	fallbackField := val.FieldByName(FALLBACK_LNG)
	if fallbackField.IsValid() && fallbackField.String() != "" {
		return fallbackField.String()
	}
	// 如果回退语言也没有消息，返回默认的错误信息
	return fmt.Sprintf("No message available for language: %s", lng)
}

// getSupportedLanguages 函数返回 lang 类型支持的所有语言
func GetSupportedLanguages() []string {
	var languages []string
	// 通过反射获取 lang 类型的字段
	typ := reflect.TypeOf(lang{})
	// 遍历结构体的字段，获取字段名
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		languages = append(languages, field.Name)
	}
	return languages
}

// 设置全局默认语言
func SetGlobalDefaultLang(language string) error {
	// 支持的语言列表
	supportedLanguages := GetSupportedLanguages()

	// 检查语言是否在支持的语言列表中
	isValidLang := false
	for _, lang := range supportedLanguages {
		if language == lang {
			isValidLang = true
			break
		}
	}
	// 如果语言有效，设置全局语言
	if isValidLang {
		lng = language
		return nil
	}
	// 如果语言无效，返回错误并设置为默认语言
	lng = FALLBACK_LNG
	return errors.New("unsupported language type, set defaulting to " + FALLBACK_LNG)
}

// 设置全局默认语言
func GetGlobalDefaultLang() string {
	return lng
}
