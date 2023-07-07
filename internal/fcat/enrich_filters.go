package fcat

import "maya/cons"

func EnrichFilters(filterMap map[string]any) []FieldCatalog {
	fieldsCatalog := []FieldCatalog{}
	for k, v := range filterMap {
		fcat := FieldCatalog{
			Name:  k,
			Tcol:  cons.FILTER,
			Value: v,
		}
		fieldsCatalog = append(fieldsCatalog, fcat)
	}
	return fieldsCatalog
}
