package fcat

import (
	"context"
	"strings"

	"github.com/edervzz/maya/cons"

	"time"
)

func EnrichAuditFieldCatalog(query string, entity any, ctx context.Context) []FieldCatalog {
	// 1.
	ctxUser := ctx.Value(cons.MAYA_USER_EDITOR)
	byUser := ""
	if ctxUser != nil {
		byUser = ctxUser.(string)
	} else {
		byUser = "local"
	}
	atTime := time.Now().UTC()
	// 1. try to confirm whether entity implements IAuditable and field catalog should be filled
	fieldCat := []FieldCatalog{}
	// 2. when query is an INSERT add creation info to field catalog
	if strings.Contains(strings.ToLower(query), "insert") {
		fieldCat = append(fieldCat, FieldCatalog{
			Name:  cons.CREATED_BY,
			Tcol:  cons.TCOL_TAG,
			Value: byUser,
		}, FieldCatalog{
			Name:  cons.CREATED_AT,
			Tcol:  cons.TCOL_TAG,
			Value: atTime,
		})
	}
	// 3. add update info
	fieldCat = append(fieldCat, FieldCatalog{
		Name:  cons.UPDATE_BY,
		Tcol:  cons.TCOL_TAG,
		Value: byUser,
	}, FieldCatalog{
		Name:  cons.UPDATE_AT,
		Tcol:  cons.TCOL_TAG,
		Value: atTime,
	})

	return fieldCat
}
