package service

import (
	"container/heap"
	"context"
	"math"
	"time"

	"github.com/tsukaychan/mercury/article/domain"

	"github.com/tsukaychan/mercury/article/service"

	interactivev1 "github.com/tsukaychan/mercury/api/proto/gen/interactive/v1"

	"github.com/ecodeclub/ekit/slice"

	"github.com/tsukaychan/mercury/internal/repository"
)

//go:generate mockgen -source=ranking.go -package=svcmocks -destination=mocks/ranking.mock.go RankingService
type RankingService interface {
	// RankTopN Calculate TopN
	RankTopN(ctx context.Context) error
	// TopN GetTopN IDs
	TopN(ctx context.Context) ([]domain.Article, error)
}

var _ RankingService = (*BatchRankingService)(nil)

type BatchRankingService struct {
	atclSvc service.ArticleService
	intrSvc interactivev1.InteractiveServiceClient

	repo      repository.RankingRepository
	BatchSize int
	TopNSize  int // limit topN Size

	scoreFunc func(likeCnt int64, utime time.Time) float64
}

func NewBatchRankingService(
	atclSvc service.ArticleService,
	intrSvc interactivev1.InteractiveServiceClient,
	repo repository.RankingRepository,
) RankingService {
	svc := &BatchRankingService{
		intrSvc:   intrSvc,
		atclSvc:   atclSvc,
		repo:      repo,
		BatchSize: 100,
		TopNSize:  200,
	}
	svc.scoreFunc = svc.score
	return svc
}

func (svc *BatchRankingService) RankTopN(ctx context.Context) error {
	atcls, err := svc.rankTopN(ctx)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceTopN(ctx, atcls)
}

type score struct {
	atcl  domain.Article
	score float64
}

type scorePriorityQueue []score

func (hp *scorePriorityQueue) Len() int {
	return len(*hp)
}

func (hp *scorePriorityQueue) Less(i, j int) bool {
	return (*hp)[i].score < (*hp)[j].score
}

func (hp *scorePriorityQueue) Swap(i, j int) {
	(*hp)[i], (*hp)[j] = (*hp)[j], (*hp)[i]
}

func (hp *scorePriorityQueue) Push(v any) {
	*hp = append(*hp, v.(score))
}

func (hp *scorePriorityQueue) Pop() any {
	v := (*hp)[len(*hp)-1]
	*hp = (*hp)[:len(*hp)-1]
	return v
}

func (hp *scorePriorityQueue) push(v score) {
	heap.Push(hp, v)
}

func (hp *scorePriorityQueue) pop() score {
	return heap.Pop(hp).(score)
}

// pop up and return to the top of the heap, while pushing v into the heap
func (hp *scorePriorityQueue) replace(v score) score {
	top := (*hp)[0]
	(*hp)[0] = v
	heap.Fix(hp, 0)
	return top
}

func (hp *scorePriorityQueue) top() score {
	return (*hp)[0]
}

func (svc *BatchRankingService) rankTopN(ctx context.Context) ([]domain.Article, error) {
	// min-heap
	topN := &scorePriorityQueue{}

	now := time.Now()
	ddl := now.Add(-time.Hour * 24 * 7)
	offset := 0

	for {
		// get a batch of publishedArticles
		atcls, err := svc.atclSvc.ListPub(ctx, now, offset, svc.BatchSize)
		if err != nil {
			return nil, err
		}

		// get ids
		atclIds := slice.Map[domain.Article, int64](atcls, func(idx int, src domain.Article) int64 {
			return src.Id
		})

		// ues ids get interactive infos from intrSvc
		resp, err := svc.intrSvc.GetByIds(ctx, &interactivev1.GetByIdsRequest{
			Biz:    "article",
			BizIds: atclIds,
		})
		if err != nil {
			return nil, err
		}

		for _, atcl := range atcls {
			intr, ok := resp.Interactives[atcl.Id]
			if !ok {
				continue
			}
			ele := score{atcl: atcl, score: svc.scoreFunc(intr.LikeCnt, atcl.Utime)}

			if topN.Len() < svc.TopNSize {
				topN.push(ele)
			} else if ele.score > topN.top().score {
				topN.replace(ele)
			}
		}

		// validate
		if len(atcls) == 0 || len(atcls) < svc.BatchSize || atcls[len(atcls)-1].Utime.Before(ddl) {
			break
		}

		// maintain offset
		offset = offset + len(atcls)
	}

	n := topN.Len()
	res := make([]domain.Article, n)
	for i := n - 1; i >= 0; i-- {
		val := topN.pop()
		res[i] = val.atcl
	}

	return res, nil
}

func (svc *BatchRankingService) TopN(ctx context.Context) ([]domain.Article, error) {
	return svc.repo.GetTopN(ctx)
}

// Algo Hacker News (p - 1) / (t + 2) ^ 1.5
func (svc *BatchRankingService) score(likeCnt int64, utime time.Time) float64 {
	const factor = 1.5
	return float64(likeCnt-1) / math.Pow(time.Since(utime).Hours()+2, factor)
}
