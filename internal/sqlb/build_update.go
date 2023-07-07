package sqlb

import (
	"context"
	"fmt"
	"maya/cons"
	"maya/internal/fcat"
	"strings"
)

// build update form => "UPDATE <tab_name> SET field1 = ?, field2 = ? WHERE fieldN = ?;" and return query + values
func BuildUpdate(ctx context.Context, entity any, isAuditable bool) (string, []any, string) {
	query := `UPDATE {TABLE_NAME} SET {FIELDS} WHERE {FILTER};`
	// 1. enrich table name, field cat, audit fields
	tableName := fcat.EnrichTableName(entity)
	fieldsCatalog := fcat.EnrichFieldsCatalog(entity)
	// var e interface{} = entity
	if isAuditable {
		auditFieldCat := fcat.EnrichAuditFieldCatalog(query, entity, ctx)
		fieldsCatalog = append(fieldsCatalog, auditFieldCat...)
	}
	// 2. prepare text for fields, question marks and values
	fields, filters := []string{}, []string{}
	values := []any{}
	filterValues := []any{}
	for _, v := range fieldsCatalog {
		field := fmt.Sprintf(" %s = ?", v.Name)
		if v.Tcol == cons.PKEY_TAG {
			filters = append(filters, field)
			filterValues = append(filterValues, v.Value)
		} else if v.Tcol == cons.TCOL_TAG {
			fields = append(fields, field)
			values = append(values, v.Value)
		}
	}
	values = append(values, filterValues...)
	// 6. replace into query, fields and filter
	query = strings.Replace(query, "{TABLE_NAME}", tableName, 1)
	query = strings.Replace(query, "{FIELDS}", strings.Join(fields, ","), 1)
	query = strings.Replace(query, "{FILTER}", strings.Join(filters, " AND "), 1)

	return query, values, tableName
}
