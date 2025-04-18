package sqlx

import (
	"fmt"
	"reflect"
	"strings"
)

// BuildInsertSQL 根据字段名生成 Insert SQL
func BuildInsertSQL(table string, fields ...string) (query string) {
	if len(fields) > 0 {
		fieldStr := strings.Join(fields, ",")
		valueStr := strings.Join(fields, ",:")
		query = fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(:%s)", table, fieldStr, valueStr)
	}
	return
}

func BuildInsertIgnoreSQL(table string, fields ...string) (query string) {
	if len(fields) > 0 {
		fieldStr := strings.Join(fields, ",")
		valueStr := strings.Join(fields, ",:")
		query = fmt.Sprintf("INSERT IGNORE INTO `%s`(%s) VALUES(:%s)", table, fieldStr, valueStr)
	}
	return
}

func GetFields(i interface{}) (fields []string) {
	_, fields = getNameAndTags(i)
	return
}

func getNameAndTags(i interface{}) (names, tags []string) {
	return getNameAndTagsT(reflect.TypeOf(i))
}

func getNameAndTagsT(t reflect.Type) (names, tags []string) {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		k := f.Type.Kind()
		if k == reflect.Struct && f.Anonymous {
			tempNames, tempTags := getNameAndTagsT(f.Type)
			tags = append(tags, tempTags...)
			names = append(names, tempNames...)
		} else {
			tag, ok := f.Tag.Lookup("db")
			if ok {
				tags = append(tags, strings.Split(tag, ",")[0])
				names = append(names, f.Name)
			}
		}
	}
	return
}
