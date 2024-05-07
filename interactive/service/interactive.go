package service

import (
	"context"

	"github.com/tsukaychan/mercury/interactive/domain"

	"github.com/tsukaychan/mercury/interactive/repository"

	"golang.org/x/sync/errgroup"

	"github.com/tsukaychan/mercury/pkg/logger"
)

//go:generate mockgen -source=interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Favorite(ctx context.Context, biz string, bizId, uid, fid int64) error
	CancelFavorite(ctx context.Context, biz string, bizId, uid, fid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.Logger
}

func NewInteractiveService(repo repository.InteractiveRepository, l logger.Logger) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    l,
	}
}

func (svc *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return svc.repo.IncrReadCnt(ctx, biz, bizId)
}

func (svc *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return svc.repo.IncrLike(ctx, biz, bizId, uid)
}

func (svc *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return svc.repo.DecrLike(ctx, biz, bizId, uid)
}

func (svc *interactiveService) Favorite(ctx context.Context, biz string, bizId, uid, fid int64) error {
	return svc.repo.AddFavoriteItem(ctx, biz, bizId, uid, fid)
}

func (svc *interactiveService) CancelFavorite(ctx context.Context, biz string, bizId, uid, fid int64) error {
	return svc.repo.DelFavoriteItem(ctx, biz, bizId, uid, fid)
}

func (svc *interactiveService) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	var (
		intr             domain.Interactive
		liked, favorited bool
		eg               errgroup.Group
	)

	eg.Go(func() error {
		var er error
		intr, er = svc.repo.Get(ctx, biz, bizId)
		return er
	})

	if uid > 0 {
		eg.Go(func() error {
			var er error
			liked, er = svc.repo.Liked(ctx, biz, bizId, uid)
			return er
		})
		eg.Go(func() error {
			var er error
			favorited, er = svc.repo.Favorited(ctx, biz, bizId, uid)
			return er
		})
	}

	err := eg.Wait()
	if err != nil {
		svc.l.Error("get user liked info failed",
			logger.String("biz", biz),
			logger.Int64("biz_id", bizId),
			logger.Int64("user_id", uid),
			logger.Error(err),
		)
	}
	intr.Liked, intr.Favorited = liked, favorited
	return intr, err
}

func (svc *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	intrs, err := svc.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}
