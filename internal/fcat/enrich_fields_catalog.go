package fcat

import (
	"maya/cons"
	"reflect"
	"strconv"
)

type FieldCatalog struct {
	Name            string
	Tcol            string
	IsAutoIncrement bool
	Value           any
}

func EnrichFieldsCatalog(entity any) []FieldCatalog {
	// 1. prepare field catalog slice and reflection
	fieldCat := []FieldCatalog{}
	valueInfo := reflect.ValueOf(entity).Elem()
	// 2. for each field entity structure
	for i := 0; i < valueInfo.NumField(); i++ {
		// 2.1 get tag for primary-key and column()
		pk := valueInfo.Type().Field(i).Tag.Get(cons.PKEY_TAG)
		colName := valueInfo.Type().Field(i).Tag.Get(cons.TCOL_TAG)
		isAutoIncrement := valueInfo.Type().Field(i).Tag.Get(cons.AUTO_INCR)
		if colName != "" {
			// 2.2 prepare field catalog for field
			fcat := FieldCatalog{
				Name:  colName,
				Value: valueInfo.Field(i).Addr().Interface(),
			}
			// 2.3 set either primary-key or column and append
			if isPKey, _ := strconv.ParseBool(pk); isPKey {
				fcat.Tcol = cons.PKEY_TAG
			} else {
				fcat.Tcol = cons.TCOL_TAG
			}
			fcat.IsAutoIncrement, _ = strconv.ParseBool(isAutoIncrement)
			fieldCat = append(fieldCat, fcat)
		}
	}

	return fieldCat
}
