package article

import "context"

type AuthorDAO interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
}
