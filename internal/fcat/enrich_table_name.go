package fcat

import (
	"maya/cons"
	"reflect"
)

func EnrichTableName(entity any) string {
	tt := reflect.TypeOf(entity).Elem()
	// 2. get table name and replace into query
	for i := 0; i < tt.NumField(); i++ {
		if tablename, ok := tt.Field(i).Tag.Lookup(cons.TABLE_NAME_TAG); ok {
			return tablename
		}
	}
	panic("entity without table name tag defined")
}
