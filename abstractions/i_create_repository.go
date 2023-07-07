package abstractions

import "context"

type ICreateRepository[TEntity any] interface {
	Create(context.Context, TEntity) error
}
