package abstractions

import "context"

type IUpdateRepository[TEntity any] interface {
	Update(context.Context, TEntity) error
}
