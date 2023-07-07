package sqlb

import (
	"context"
	"fmt"
	"strings"

	"github.com/edervzz/maya/internal/fcat"
)

func BuildRead(ctx context.Context, entity any, filter map[string]any) (string, []any, string) {
	query := `SELECT {FIELDS} FROM {TABLE} {CONDITIONS};`
	tableName := fcat.EnrichTableName(entity)
	fieldsCatalog := fcat.EnrichFieldsCatalog(entity)
	filtersCatalog := fcat.EnrichFilters(filter)
	// prepare field catalog
	fields := []string{}
	for _, v := range fieldsCatalog {
		fields = append(fields, fmt.Sprintf("%s", v.Name))
	}
	// prepare field catalog for filter
	filters := []string{}
	values := []any{}
	where := ""
	for _, v := range filtersCatalog {
		filters = append(filters, fmt.Sprintf("%s = ?", v.Name))
		values = append(values, v.Value)
	}
	if len(filters) > 0 {
		where = "WHERE " + strings.Join(filters, " AND ")
	}

	query = strings.Replace(query, "{TABLE}", tableName, 1)
	query = strings.Replace(query, "{FIELDS}", strings.Join(fields, ", "), 1)
	query = strings.Replace(query, "{CONDITIONS}", where, 1)
	return query, values, tableName
}
