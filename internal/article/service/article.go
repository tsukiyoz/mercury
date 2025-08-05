package service

import (
	"context"
	"time"

	userv1 "github.com/tsukiyo/mercury/api/gen/user/v1"

	"golang.org/x/sync/errgroup"

	"github.com/tsukiyo/mercury/internal/article/domain"
	"github.com/tsukiyo/mercury/internal/article/events"
	"github.com/tsukiyo/mercury/internal/article/repository"
	"github.com/tsukiyo/mercury/pkg/logger"
)

var _ ArticleService = (*articleService)(nil)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=mocks/article.mock.go ArticleService
type ArticleService interface {
	// author

	Save(ctx context.Context, atcl domain.Article) (int64, error)
	Publish(ctx context.Context, atcl domain.Article) (int64, error)
	Withdraw(ctx context.Context, id, authorId int64) error
	List(ctx context.Context, authorId int64,
		offset, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)

	// reader

	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	articleRepo repository.ArticleRepository
	userSvc     userv1.UserServiceClient
	producer    events.Producer
	logger      logger.Logger
}

func NewArticleService(
	articleRepo repository.ArticleRepository,
	userSvc userv1.UserServiceClient,
	producer events.Producer,
	logger logger.Logger,
) ArticleService {
	return &articleService{
		articleRepo: articleRepo,
		userSvc:     userSvc,
		producer:    producer,
		logger:      logger,
	}
}

func (svc *articleService) Save(ctx context.Context, atcl domain.Article) (int64, error) {
	atcl.Status = domain.ArticleStatusUnpublished
	if atcl.Id > 0 {
		err := svc.articleRepo.Update(ctx, atcl)
		return atcl.Id, err
	}

	return svc.articleRepo.Create(ctx, atcl)
}

func (svc *articleService) Publish(ctx context.Context, atcl domain.Article) (int64, error) {
	atcl.Status = domain.ArticleStatusPublished
	author, err := svc.findAuthor(ctx, atcl.Id)
	if err != nil {
		return 0, err
	}
	atcl.Author = author
	return svc.articleRepo.Sync(ctx, atcl)
}

func (svc *articleService) Withdraw(ctx context.Context, id, authorId int64) error {
	return svc.articleRepo.SyncStatus(ctx, id, authorId, domain.ArticleStatusPrivate)
}

func (svc *articleService) List(ctx context.Context, authorId int64, offset, limit int) ([]domain.Article, error) {
	return svc.articleRepo.List(ctx, authorId, offset, limit)
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.articleRepo.GetById(ctx, id)
}

func (svc *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	var eg errgroup.Group
	var err error
	var atcl *domain.Article
	var author *domain.Author
	eg.Go(func() error {
		res, err := svc.articleRepo.GetPublishedById(ctx, id)
		atcl = &res
		return err
	})
	eg.Go(func() error {
		res, err := svc.findAuthor(ctx, id)
		author = &res
		return err
	})
	err = eg.Wait()
	if err != nil {
		return domain.Article{}, err
	}
	atcl.Author = *author
	if err == nil {
		go func() {
			er := svc.producer.ProduceReadEvent(events.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				svc.logger.Error("send reader read event failed",
					logger.Int64("uid", uid),
					logger.Int64("aid", id),
					logger.Error(err))
			}
		}()
	}
	return *atcl, err
}

func (svc *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return svc.articleRepo.ListPub(ctx, start, offset, limit)
}

func (svc *articleService) findAuthor(ctx context.Context, id int64) (domain.Author, error) {
	atcl, err := svc.articleRepo.GetPublishedById(ctx, id)
	if err != nil {
		return domain.Author{}, err
	}
	resp, err := svc.userSvc.Profile(ctx, &userv1.ProfileRequest{
		Id: atcl.Author.Id,
	})
	if err != nil {
		return domain.Author{}, err
	}
	return domain.Author{
		Id:   resp.GetUser().GetId(),
		Name: resp.GetUser().GetNickName(),
	}, nil
}
