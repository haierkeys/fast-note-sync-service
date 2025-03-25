package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/haierkeys/obsidian-better-sync-service/cmd/old_gen/db_gen/db_driver"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/convert"
)

type tableInfo struct {
	Name    string         `db_driver:"table_name"`    // name
	Comment sql.NullString `db_driver:"table_comment"` // comment
}

type tableColumn struct {
	OrdinalPosition uint16         `db_driver:"ORDINAL_POSITION"` // position
	ColumnName      string         `db_driver:"COLUMN_NAME"`      // name
	ColumnType      string         `db_driver:"COLUMN_TYPE"`      // column_type
	DataType        string         `db_driver:"DATA_TYPE"`        // data_type
	ColumnKey       sql.NullString `db_driver:"COLUMN_KEY"`       // key
	IsNullable      string         `db_driver:"IS_NULLABLE"`      // nullable
	Extra           sql.NullString `db_driver:"EXTRA"`            // extra
	ColumnComment   sql.NullString `db_driver:"COLUMN_COMMENT"`   // comment
	ColumnDefault   sql.NullString `db_driver:"COLUMN_DEFAULT"`   // default value
}

var (
	dbType         string
	dbDsn          string
	dbName         string
	genTablePrefix string
	tablePrefix    string
	repoSaveDir    string
)

func init() {
	dType := flag.String("type", "", "输入类型")
	dsn := flag.String("dsn", "", "输入DB dsn地址")
	name := flag.String("name", "", "输入DB name")
	table := flag.String("table", "", "请输入需要处理的数据表名前缀，默认为所有\n")
	prefix := flag.String("prefix", "", "请输入 prefix 名称\n")
	saveDir := flag.String("savedir", "", "请输入 保存目录\n")

	flag.Parse()

	dbType = *dType
	dbDsn = *dsn
	dbName = *name

	genTablePrefix = strings.ToLower(*table)
	tablePrefix = *prefix
	repoSaveDir = *saveDir

}

func main() {

	var db db_driver.Repo
	var err error

	if dbType == "mysql" {

		// 初始化 DB
		db, err = db_driver.NewMysql(dbDsn)
		if err != nil {
			log.Fatal("new db_driver err", err)
		}
	} else if dbType == "sqlite" {

		// 初始化 DB
		db, err = db_driver.NewSqlite(dbDsn)
		if err != nil {
			log.Fatal("new db_driver err", err)
		}
	} else {
		flag.Usage()
		return
	}

	defer func() {
		if err := db.DbClose(); err != nil {
			log.Println("db_driver close err", err)
		}
	}()

	tables, err := queryTables(db, dbName, genTablePrefix)
	if err != nil {
		log.Println("query tables of database err", err)
		return
	}

	var tableSaveName string

	for _, table := range tables {

		if tablePrefix != "" {
			tableSaveName = strings.Replace(table.Name, tablePrefix, "", 1)
		} else {
			tableSaveName = table.Name
		}

		var dbPath string
		if repoSaveDir != "" {
			dbPath = "./internal/model/" + repoSaveDir
			_ = os.Mkdir(dbPath, 0766)
		} else {
			dbPath = "./internal/model/"
		}

		filepath := dbPath + "/" + tableSaveName + "_repo"
		_ = os.Mkdir(filepath, 0766)
		fmt.Println("create dir : ", filepath)

		mdName := fmt.Sprintf("%s/table.md", filepath)
		mdFile, err := os.OpenFile(mdName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0766)
		if err != nil {
			fmt.Printf("markdown file error %v\n", err.Error())
			return
		}
		fmt.Println("  └── file : ", dbPath+"/"+tableSaveName+"_repo/table.md")

		modelName := fmt.Sprintf("%s/model.go", filepath)
		modelFile, err := os.OpenFile(modelName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0766)
		if err != nil {
			fmt.Printf("create and open model file error %v\n", err.Error())
			return
		}
		fmt.Println("  └── file : ", dbPath+"/"+tableSaveName+"_repo/model.go")

		modelContent := fmt.Sprintf("package %s%s\n", tableSaveName, "_repo")
		modelContent += fmt.Sprintf(`import "github.com/haierkeys/obsidian-better-sync-service/pkg/timex"`)
		modelContent += fmt.Sprintf("\n\n// %s \n", table.Comment.String)
		modelContent += fmt.Sprintf("//go:generate gormgen -structs %s -input . -pre %s \n", capitalize(tableSaveName), tablePrefix)
		modelContent += fmt.Sprintf("type %s struct {\n", capitalize(tableSaveName))

		tableContent := fmt.Sprintf("#### %s.%s \n", dbName, tableSaveName)
		if table.Comment.String != "" {
			tableContent += table.Comment.String + "\n"
		}
		tableContent += "\n" +
			"| 序号 | 名称 | 描述 | 类型 | 键 | 为空 | 额外 | 默认值 |\n" +
			"| :--: | :--: | :--: | :--: | :--: | :--: | :--: | :--: |\n"

		columnInfo, columnInfoErr := queryTableColumn(db, dbName, table.Name)
		if columnInfoErr != nil {
			continue
		}
		for _, info := range columnInfo {
			tableContent += fmt.Sprintf(
				"| %d | %s | %s | %s | %s | %s | %s | %s |\n",
				info.OrdinalPosition,
				info.ColumnName,
				strings.ReplaceAll(strings.ReplaceAll(info.ColumnComment.String, "|", "\\|"), "\n", ""),
				info.ColumnType,
				info.ColumnKey.String,
				info.IsNullable,
				info.Extra.String,
				info.ColumnDefault.String,
			)

			gormAdd := []string{}

			gormAdd = append(gormAdd, "column:"+info.ColumnName)

			if info.ColumnKey.Valid && info.ColumnKey.String == "PRI" {
				gormAdd = append(gormAdd, "primary_key")
			}
			if info.ColumnKey.Valid && info.ColumnKey.String == "INDEX" {
				gormAdd = append(gormAdd, "index")
			}
			if info.Extra.Valid && info.Extra.String == "auto_increment" {
				gormAdd = append(gormAdd, "auto_increment")
			}
			if textType(info.DataType, db.DbType()) == "timex.Time" {
				gormAdd = append(gormAdd, "type:datetime")
				gormAdd = append(gormAdd, "autoUpdateTime:false")
			}

			if info.ColumnDefault.Valid {
				gormAdd = append(gormAdd, "default:"+strings.ReplaceAll(info.ColumnDefault.String, "\"", "'"))
			}

			jsonAdd := convert.Case2LowerCamel(info.ColumnName)
			formAdd := convert.Case2LowerCamel(info.ColumnName)

			if len(gormAdd) > 0 {
				modelContent += fmt.Sprintf("%s %s `%s %s %s` // %s\n", capitalize(info.ColumnName), textType(info.DataType, db.DbType()), "gorm:\""+strings.Join(gormAdd, ";")+"\"", "json:\""+jsonAdd+"\"", "form:\""+formAdd+"\"", info.ColumnComment.String)
			} else {
				modelContent += fmt.Sprintf("%s %s // %s\n", capitalize(info.ColumnName), textType(info.DataType, db.DbType()), info.ColumnComment.String)
			}

		}

		mdFile.WriteString(tableContent)
		mdFile.Close()

		modelContent += "}\n"
		modelFile.WriteString(modelContent)
		modelFile.Close()

	}

}

func queryTables(dbIns db_driver.Repo, dbName string, genTablePrefix string) ([]tableInfo, error) {
	var tableCollect []tableInfo
	var tableArray []string
	var commentArray []sql.NullString

	db := dbIns.GetDb()
	var sqlTables string

	if dbIns.DbType() == "mysql" {
		sqlTables = fmt.Sprintf("SELECT `table_name`,`table_comment` FROM `information_schema`.`tables` WHERE `table_schema`= '%s'", dbName)
	} else if dbIns.DbType() == "sqlite" {
		sqlTables = fmt.Sprint("select name from sqlite_master where type='table' order by name;")
	}

	rows, err := db.Raw(sqlTables).Rows()

	if err != nil {
		return tableCollect, err
	}
	defer rows.Close()

	for rows.Next() {
		var info tableInfo
		if dbIns.DbType() == "mysql" {
			err = rows.Scan(&info.Name, &info.Comment)
		} else if dbIns.DbType() == "sqlite" {
			err = rows.Scan(&info.Name)
		}

		if err != nil {
			fmt.Printf("execute query tables action error,had ignored, detail is [%v]\n", err.Error())
			continue
		}

		if !strings.HasPrefix(info.Name, genTablePrefix) {
			continue
		}

		tableCollect = append(tableCollect, info)
		tableArray = append(tableArray, info.Name)
		commentArray = append(commentArray, info.Comment)
	}

	return tableCollect, err
}

func queryTableColumn(dbIns db_driver.Repo, dbName string, tableName string) ([]tableColumn, error) {

	db := dbIns.GetDb()

	// 定义承载列信息的切片
	var columns []tableColumn
	var indexs = make(map[string]bool)
	var sqlTableColumn string

	if dbIns.DbType() == "mysql" {
		sqlTableColumn = fmt.Sprintf("SELECT `ORDINAL_POSITION`,`COLUMN_NAME`,`COLUMN_TYPE`,`DATA_TYPE`,`COLUMN_KEY`,`IS_NULLABLE`,`EXTRA`,`COLUMN_COMMENT`,`COLUMN_DEFAULT` FROM `information_schema`.`columns` WHERE `table_schema`= '%s' AND `table_name`= '%s' ORDER BY `ORDINAL_POSITION` ASC",
			dbName, tableName)
	} else if dbIns.DbType() == "sqlite" {
		sqlTableColumn = fmt.Sprintf("PRAGMA table_info(%s);", tableName)

		sqlAllIndex := fmt.Sprintf("SELECT name AS 'index' FROM sqlite_master WHERE type = 'index' AND tbl_name = '%s'", tableName)
		rowsAI, err := db.Raw(sqlAllIndex).Rows()
		if err != nil {
			log.Fatal(err)
		}
		defer rowsAI.Close()

		for rowsAI.Next() {
			var tableIndex string
			err = rowsAI.Scan(&tableIndex)
			if tableIndex != "" {
				sqli := fmt.Sprintf("SELECT name AS 'index_name' FROM pragma_index_info('%s')", tableIndex)
				rowsi, err := db.Raw(sqli).Rows()
				defer rowsi.Close()

				for rowsi.Next() {
					var name string
					err = rowsi.Scan(&name)
					if err == nil {
						indexs[name] = true
					}
				}
			}

		}
	}

	rows, err := db.Raw(sqlTableColumn).Rows()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {

		var column tableColumn
		var err error

		if dbIns.DbType() == "mysql" {
			err = rows.Scan(
				&column.OrdinalPosition,
				&column.ColumnName,
				&column.ColumnType,
				&column.DataType,
				&column.ColumnKey,
				&column.IsNullable,
				&column.Extra,
				&column.ColumnComment,
				&column.ColumnDefault)

		} else if dbIns.DbType() == "sqlite" {

			var pk int

			err = rows.Scan(
				&column.OrdinalPosition,
				&column.ColumnName,
				&column.ColumnType,
				&column.IsNullable,
				&column.ColumnDefault, &pk)

			if pk == 1 {
				column.ColumnKey = sql.NullString{
					String: "PRI",
					Valid:  true,
				}
				column.Extra = sql.NullString{
					String: "auto_increment",
					Valid:  true,
				}
			} else {
				if indexs[column.ColumnName] {
					column.ColumnKey = sql.NullString{
						String: "INDEX",
						Valid:  true,
					}
				}
				column.Extra = sql.NullString{
					String: "",
					Valid:  true,
				}

			}

			column.DataType = column.ColumnType
		}

		if err != nil {
			fmt.Printf("query table column scan error, detail is [%v]\n", err.Error())
			return columns, err
		}

		columns = append(columns, column)
	}

	return columns, err
}

func capitalize(s string) string {
	var upperStr string
	chars := strings.Split(s, "_")
	for _, val := range chars {
		vv := []rune(val)
		for i := 0; i < len(vv); i++ {
			if i == 0 {
				if vv[i] >= 97 && vv[i] <= 122 {
					vv[i] -= 32
					upperStr += string(vv[i])
				}
			} else {
				upperStr += string(vv[i])
			}
		}
	}
	return upperStr
}

func textType(s string, t string) string {

	s = strings.ToLower(s)
	var sqliteTypeToGoType = map[string]string{
		"integer":   "int64",
		"text":      "string",
		"timestamp": "timex.Time",
		"datetime":  "timex.Time",
		"real":      "float64",
	}
	var mysqlTypeToGoType = map[string]string{
		"tinyint":    "int32",
		"smallint":   "int32",
		"mediumint":  "int32",
		"int":        "int32",
		"integer":    "int64",
		"bigint":     "int64",
		"float":      "float64",
		"double":     "float64",
		"decimal":    "float64",
		"date":       "string",
		"time":       "string",
		"year":       "string",
		"datetime":   "timex.Time",
		"timestamp":  "timex.Time",
		"char":       "string",
		"varchar":    "string",
		"tinyblob":   "string",
		"tinytext":   "string",
		"blob":       "string",
		"text":       "string",
		"mediumblob": "string",
		"mediumtext": "string",
		"longblob":   "string",
		"longtext":   "string",
	}
	if t == "sqlite" {
		return sqliteTypeToGoType[s]
	} else if t == "mysql" {
		return mysqlTypeToGoType[s]
	} else {
		return mysqlTypeToGoType[s]
	}

}
