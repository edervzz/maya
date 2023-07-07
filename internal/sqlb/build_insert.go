package sqlb

import (
	"context"
	"fmt"
	"maya/internal/fcat"
	"strings"
)

// build insert form => "INSERT INTO <tab_name> (field1, field2, fieldN) VALUES(?,?,?);" and return query + values
func BuildInsert(ctx context.Context, entity any, isAuditable bool) (string, []any, string) {
	query := `INSERT INTO {TABLE_NAME} ({FIELDS}) VALUES({VALUES});`
	// 1. enrich table name, field cat, audit fields
	tableName := fcat.EnrichTableName(entity)
	fieldsCatalog := fcat.EnrichFieldsCatalog(entity)
	if isAuditable {
		auditFieldCat := fcat.EnrichAuditFieldCatalog(query, entity, ctx)
		fieldsCatalog = append(fieldsCatalog, auditFieldCat...)
	}
	// 2. prepare text for fields, question marks and values
	fields, qm := []string{}, []string{}
	values := []any{}
	for _, v := range fieldsCatalog {
		if v.IsAutoIncrement {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s", v.Name))
		qm = append(qm, "?")
		values = append(values, v.Value)
	}
	// 3. replace into query, fields and values
	query = strings.Replace(query, "{TABLE_NAME}", tableName, 1)
	query = strings.Replace(query, "{FIELDS}", strings.Join(fields, ","), 1)
	query = strings.Replace(query, "{VALUES}", strings.Join(qm, ","), 1)

	return query, values, tableName
}
