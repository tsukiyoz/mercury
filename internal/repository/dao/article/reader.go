package article

import (
	"context"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, article Article) error
}

type PublishArticle struct {
	Article
}
