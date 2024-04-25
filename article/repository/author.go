package repository

import (
	"context"

	userv1 "github.com/tsukaychan/mercury/api/proto/gen/user/v1"
	"github.com/tsukaychan/mercury/article/domain"
	"github.com/tsukaychan/mercury/article/repository/dao"
)

type AuthorRepository interface {
	// FindAuthor find author by article_id
	FindAuthor(ctx context.Context, id int64) (domain.Author, error)
}

var _ AuthorRepository = (*GrpcAuthorRepository)(nil)

type GrpcAuthorRepository struct {
	client userv1.UserServiceClient
	dao    dao.ArticleDAO
}

func (g *GrpcAuthorRepository) FindAuthor(ctx context.Context, id int64) (domain.Author, error) {
	article, err := g.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Author{}, err
	}
	resp, err := g.client.Profile(ctx, &userv1.ProfileRequest{Id: article.AuthorId})
	if err != nil {
		return domain.Author{}, err
	}
	return domain.Author{
		Id:   resp.GetUser().GetId(),
		Name: resp.GetUser().GetNickName(),
	}, nil
}

func NewGrpcAuthorRepository(articleDao dao.ArticleDAO, client userv1.UserServiceClient) AuthorRepository {
	return &GrpcAuthorRepository{
		client: client,
		dao:    articleDao,
	}
}
