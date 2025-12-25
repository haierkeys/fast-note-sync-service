package convert

import (
	"reflect"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

func GetCopyStructFields(source interface{}, target interface{}) []interface{} {
	// 1. 调用您的第一个函数获取 source 的字段名列表
	sourceFieldsList := GetStructFieldNames(source)
	if len(sourceFieldsList) == 0 {
		return nil
	}

	// 2. 将列表转为 map，方便 O(1) 复杂度查找
	sourceFieldsMap := make(map[string]bool)
	for _, name := range sourceFieldsList {
		sourceFieldsMap[name] = true
	}

	// 3. 准备提取 target 中的值
	var result []interface{}
	tVal := reflect.ValueOf(target)

	// 处理 target 的指针
	if tVal.Kind() == reflect.Ptr {
		tVal = tVal.Elem()
	}

	// 确保 target 是结构体
	if tVal.Kind() != reflect.Struct {
		return nil
	}

	tTyp := tVal.Type()
	for i := 0; i < tVal.NumField(); i++ {
		fieldName := tTyp.Field(i).Name
		// 如果 target 的这个字段名在 source 中也存在
		if sourceFieldsMap[fieldName] {
			// 获取字段的值
			fieldVal := tVal.Field(i)

			// 只有导出字段才能调用 Interface()，否则会 panic
			if fieldVal.CanInterface() {
				result = append(result, fieldVal.Interface())
			}
		}
	}

	return result
}

// GetStructFields 返回传入结构体的所有字段名
func GetStructFieldNames(input interface{}) []string {
	getType := reflect.TypeOf(input)

	// 如果传入的是指针，获取其指向的元素类型
	if getType.Kind() == reflect.Ptr {
		getType = getType.Elem()
	}

	// 确保最终处理的是结构体
	if getType.Kind() != reflect.Struct {
		return nil
	}

	fields := make([]string, 0, getType.NumField())
	for i := 0; i < getType.NumField(); i++ {
		field := getType.Field(i)
		// 如果只想获取导出字段（大写开头的），可以直接添加
		// 如果需要处理嵌套结构体或匿名首元，可在此添加逻辑
		fields = append(fields, field.Name)
	}

	return fields
}

// CopyStruct
// dst 目标结构体，src 源结构体
// 它会把src与dst的相同字段名的值，复制到dst中
func StructAssign2(src any, dst any) any {

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

// CopyStruct
// dst 目标结构体，src 源结构体
// 它会把src与dst的相同字段名的值，复制到dst中
func StructAssign(src any, dst any) any {
	copier.Copy(dst, src)
	return dst
}

/**
 * @Description: 结构体map互转
 * @param param interface{} 需要被转的数据
 * @param data interface{} 转换完成后的数据  需要用引用传进来
 * @return []string{}
 */
func StructToMap(param any, data map[string]interface{}) error {
	str, _ := sonic.Marshal(param)
	error := sonic.Unmarshal(str, &data)
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
