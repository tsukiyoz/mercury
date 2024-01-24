package article

import (
	"context"
	"github.com/tsukaychan/webook/internal/domain"
)

type ArticleReaderRepository interface {
	// Save Upsert semantics
	Save(ctx context.Context, article domain.Article) (int64, error)
}
