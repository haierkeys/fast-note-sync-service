package main

// gorm gen configure
import (
	"os"
	"reflect"
	"strings"

	"github.com/haierkeys/obsidian-better-sync-service/internal/query"
	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		// 默认会在 OutPath 目录生成CRUD代码，并且同目录下生成 model 包
		// 所以OutPath最终package不能设置为model，在有数据库表同步的情况下会产生冲突
		// 若一定要使用可以通过ModelPkgPath单独指定model package的名称
		OutPath: "./internal/query",
		/* ModelPkgPath: "dal/model"*/
		// gen.WithoutContext：禁用WithContext模式
		// gen.WithDefaultQuery：生成一个全局Query对象Q
		// gen.WithQueryInterface：生成Query接口
		Mode:             gen.WithQueryInterface,
		WithUnitTest:     false,
		FieldWithTypeTag: false,
	})
	v := reflect.ValueOf(query.Query{})
	goContent := `
package model

import (
	"gorm.io/gorm"
)



func AutoMigrate(db *gorm.DB, key string) {
	switch key {
`
	goContentFunc := `
	case "{NAME}":
		db.AutoMigrate({NAME}{})
`

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.Name == "db" {
				continue
			}
			goContent += strings.ReplaceAll(goContentFunc, "{NAME}", field.Name)
			//goContentHeader += fmt.Sprintf("type %s = %s\n", field.Name, field.Type.Name())
		}
		goContent += "\t}\n}"

		_ = os.WriteFile(g.OutPath[0:len(g.OutPath)-6]+"/model/model.go", []byte(goContent), os.ModePerm)
	}
}
