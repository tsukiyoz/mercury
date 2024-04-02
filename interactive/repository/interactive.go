package repository

import (
	"context"

	"github.com/tsukaychan/webook/interactive/domain"

	"github.com/tsukaychan/webook/interactive/repository/cache"
	"github.com/tsukaychan/webook/interactive/repository/dao"

	"github.com/ecodeclub/ekit/slice"

	"github.com/tsukaychan/webook/pkg/logger"
)

//go:generate mockgen -source=./interactive.go -package=repomocks -destination=mocks/interactive.mock.go InteractiveRepository
type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context,
		biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, biz string, bizIds []int64) error
	IncrLike(ctx context.Context, biz string, bizId, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId, uid int64) error
	AddFavoriteItem(ctx context.Context, biz string, bizId, uid int64, fid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Favorited(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

var _ InteractiveRepository = (*CachedInteractiveRepository)(nil)

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.Logger
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache, l logger.Logger) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (repo *CachedInteractiveRepository) entityToDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		Biz:         intr.Biz,
		BizId:       intr.BizId,
		ReadCnt:     intr.ReadCnt,
		LikeCnt:     intr.LikeCnt,
		FavoriteCnt: intr.FavoriteCnt,
	}
}

func (repo *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := repo.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, biz string, bizIds []int64) error {
	err := repo.dao.BatchIncrReadCnt(ctx, biz, bizIds)
	if err != nil {
		return err
	}
	return repo.cache.BatchIncrReadCntIfPresent(ctx, biz, bizIds)
}

func (repo *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId, uid int64) error {
	err := repo.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId, uid int64) error {
	err := repo.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return repo.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) AddFavoriteItem(ctx context.Context, biz string, bizId, uid int64, fid int64) error {
	err := repo.dao.InsertFavoriteItem(ctx, dao.FavoriteItem{
		Biz:   biz,
		BizId: bizId,
		Uid:   uid,
		Fid:   fid,
	})
	if err != nil {
		return err
	}
	return repo.cache.IncrFavoriteCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	intr, err := repo.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}
	dbIntr, err := repo.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	intr = repo.entityToDomain(dbIntr)
	go func() {
		er := repo.cache.Set(ctx, biz, bizId, intr)
		if er != nil {
			repo.l.Error("write back redis failed",
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Error(er),
			)
		}
	}()
	return intr, err
}

func (repo *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := repo.dao.GetLikeInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (repo *CachedInteractiveRepository) Favorited(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := repo.dao.GetFavoriteInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (repo *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := repo.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Interactive, domain.Interactive](intrs, func(idx int, src dao.Interactive) domain.Interactive {
		return repo.entityToDomain(src)
	}), nil
}
