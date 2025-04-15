package convert

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// CopyStruct
// dst 目标结构体，src 源结构体
// 它会把src与dst的相同字段名的值，复制到dst中
func StructAssign(src any, dst any) any {

	bVal := reflect.ValueOf(dst).Elem() // 获取reflect.Type类型
	vVal := reflect.ValueOf(src).Elem() // 获取reflect.Type类型
	vTypeOfT := vVal.Type()
	for i := 0; i < vVal.NumField(); i++ {
		// 在要修改的结构体中查询有数据结构体中相同属性的字段，有则修改其值
		name := vTypeOfT.Field(i).Name
		if ok := bVal.FieldByName(name).IsValid(); ok {
			bVal.FieldByName(name).Set(reflect.ValueOf(vVal.Field(i).Interface()))
		}
	}

	return dst
}

/**
 * @Description: 结构体map互转
 * @param param interface{} 需要被转的数据
 * @param data interface{} 转换完成后的数据  需要用引用传进来
 * @return []string{}
 */
func StructToMap(param any, data map[string]interface{}) error {
	str, _ := json.Marshal(param)
	error := json.Unmarshal(str, &data)
	if error != nil {
		return error
	} else {
		return nil
	}

}

func StructToMapByReflect(obj any) map[string]any {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]any)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		fieldName := typ.Field(i).Name
		if field.CanInterface() {
			// 如果字段是 Struct，递归处理
			if field.Kind() == reflect.Struct {
				result[fieldName] = StructToMapByReflect(field.Interface())
			} else {
				result[fieldName] = field.Interface()
			}
		}
	}

	return result
}

func StructToModelMap(param any, data map[string]any, key string) error {

	// 获取反射值
	val := reflect.ValueOf(param)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 确保传入的是结构体
	if val.Kind() != reflect.Struct {
		return errors.New("not struct")
	}

	// 获取结构体类型
	typ := val.Type()

	// 遍历结构体字段
	for i := 0; i < val.NumField(); i++ {

		if key == "" || typ.Field(i).Name == key {
			continue
		}

		// 获取 GORM 的 column 标签
		tags := splitGormTag(typ.Field(i).Tag.Get("gorm"))

		if tags["column"] != "" {
			data[tags["column"]] = val.Field(i).Interface()
		}
	}

	return nil

}

// 分割 GORM 标签
func splitGormTag(tag string) map[string]string {
	tags := strings.Split(tag, ";")

	parts := make(map[string]string, 0)

	for _, part := range tags {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}

	return parts
}
