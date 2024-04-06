package maya

import (
	"context"
	"database/sql"
)

type IDbContext interface {
	BeginTransaction(ctx context.Context) error
	Commit() error
	Rollback() error
	Add(ctx context.Context, entity any) (sql.Result, error)
	Update(ctx context.Context, entity any) (sql.Result, error)
	Read(ctx context.Context, entityShape any, filter map[string]any) (*sql.Rows, error)
	Enqueue(ctx context.Context, id string, entity any, info string) error
	Dequeue(ctx context.Context) error
	Sql(ctx context.Context, query string, args ...any) error
	Migrate() error
}
