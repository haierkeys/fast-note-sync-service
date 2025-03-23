package main

// gorm gen configure

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	dbType string
	dbDsn  string
)

func init() {

	dType := flag.String("type", "", "输入类型")
	dsn := flag.String("dsn", "", "输入DB dsn地址")

	flag.Parse()
	dbType = *dType
	dbDsn = *dsn
}

func main() {

	// 指定生成代码的具体相对目录(相对当前文件)，默认为：./query
	// 默认生成需要使用WithContext之后才可以查询的代码，但可以通过设置gen.WithoutContext禁用该模式
	g := gen.NewGenerator(gen.Config{
		// 默认会在 OutPath 目录生成CRUD代码，并且同目录下生成 model 包
		// 所以OutPath最终package不能设置为model，在有数据库表同步的情况下会产生冲突
		// 若一定要使用可以通过ModelPkgPath单独指定model package的名称
		OutPath: "./internal/query",
		/* ModelPkgPath: "dal/model"*/

		// gen.WithoutContext：禁用WithContext模式
		// gen.WithDefaultQuery：生成一个全局Query对象Q
		// gen.WithQueryInterface：生成Query接口
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
		WithUnitTest: true,
		//FieldWithTypeTag: true,
	})

	// 通常复用项目中已有的SQL连接配置db(*gorm.DB)
	// 非必需，但如果需要复用连接时的gorm.Config或需要连接数据库同步表信息则必须设置

	db := NewDBEngine(dbDsn, dbType)
	g.UseDB(db)

	var dataMap = map[string]func(gorm.ColumnType) (dataType string){
		// int mapping
		"integer": func(columnType gorm.ColumnType) (dataType string) {
			if n, ok := columnType.Nullable(); ok && n {
				return "int64"
			}
			return "int64"
		},
		"int": func(columnType gorm.ColumnType) (dataType string) {
			if n, ok := columnType.Nullable(); ok && n {
				return "int64"
			}
			return "int64"
		},
	}
	g.WithDataTypeMap(dataMap)

	opts := []gen.ModelOpt{
		//gen.FieldType("uid", "int64"),
		gen.FieldType("created_at", "timex.Time"),
		gen.FieldType("updated_at", "timex.Time"),
		gen.FieldType("deleted_at", "timex.Time"),
		gen.FieldGORMTag("created_at", func(tag field.GormTag) field.GormTag {
			tag.Set("type", "datetime")
			tag.Set("autoUpdateTime", "false")
			tag.Set("default", "NULL")
			return tag
		}),
		gen.FieldGORMTag("updated_at", func(tag field.GormTag) field.GormTag {
			tag.Set("type", "datetime")
			tag.Set("autoUpdateTime", "false")
			tag.Set("default", "NULL")
			return tag
		}),
		gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
			tag.Set("type", "datetime")
			tag.Set("autoUpdateTime", "false")
			tag.Set("default", "NULL")
			return tag
		}),
		gen.FieldJSONTagWithNS(func(columnName string) string {
			return SQLColumnToHumpStyle(columnName)
		}),

		gen.FieldNewTagWithNS("form", func(columnName string) string {
			return SQLColumnToHumpStyle(columnName)
		}),
		gen.FieldNewTagWithNS("type", func(columnName string) string {
			return SQLColumnToHumpStyle(columnName)
		}),
	}

	tableList, _ := db.Migrator().GetTables()

	for _, table := range tableList {
		if table == "sqlite_sequence" {
			continue
		}
		g.ApplyBasic(g.GenerateModel(table, opts...))
	}

	//g.ApplyBasic(g.GenerateAllTable()...)

	// g.GenerateModel("user", gen.WithMethod(CommonMethod{}))

	// g.GenerateModel("user", gen.WithMethod(CommonMethod{}))

	//g.ApplyBasic(g.GenerateAllTable(opts...)...)

	// 执行并生成代码
	g.Execute()
}

// SQLColumnToHumpStyle sql转换成驼峰模式
func SQLColumnToHumpStyle(in string) (ret string) {
	for i := 0; i < len(in); i++ {
		if i > 0 && in[i-1] == '_' && in[i] != '_' {
			s := strings.ToUpper(string(in[i]))
			ret += s
		} else if in[i] == '_' {
			continue
		} else {
			ret += string(in[i])
		}
	}
	return
}

func NewDBEngine(dsn string, dbType string) *gorm.DB {

	db, err := gorm.Open(userDialector(dsn, dbType), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
	})
	if err != nil {
		panic(fmt.Errorf("connect db fail: %w", err))
	}
	return db
}

func userDialector(dsn string, dbType string) gorm.Dialector {
	if dbType == "mysql" {
		return mysql.Open(dsn)
	} else if dbType == "sqlite" {

		if !fileurl.IsExist(dsn) {
			fileurl.CreatePath(dsn, os.ModePerm)
		}
		return sqlite.Open(dsn)
	}
	return nil
}
