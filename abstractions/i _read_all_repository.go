package abstractions

import "context"

type IReadAllRepository[TEntity any] interface {
	ReadAll(context.Context) ([]TEntity, error)
}
