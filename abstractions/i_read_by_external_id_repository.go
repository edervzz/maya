package abstractions

import "context"

type IReadByExternalIDRepository[TKey any, TEntity any] interface {
	ReadByExternalID(context.Context, TKey) (TEntity, error)
}
