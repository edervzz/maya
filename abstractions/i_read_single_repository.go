package abstractions

import "context"

type IReadSingleRepository[TKey any, TEntity any] interface {
	Read(TKey, context.Context) (TEntity, error)
}
