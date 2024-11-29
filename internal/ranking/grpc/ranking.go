package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/lazywoo/mercury/internal/article/domain"

	"google.golang.org/grpc"

	rankingv1 "github.com/lazywoo/mercury/api/gen/ranking/v1"
	"github.com/lazywoo/mercury/internal/ranking/service"
)

type RankingServiceServer struct {
	rankingv1.RankingServiceServer

	svc service.RankingService
}

func NewRankingServiceServer(svc service.RankingService) *RankingServiceServer {
	return &RankingServiceServer{
		svc: svc,
	}
}

func (r *RankingServiceServer) Register(server grpc.ServiceRegistrar) {
	rankingv1.RegisterRankingServiceServer(server, r)
}

func (r *RankingServiceServer) RankTopN(ctx context.Context, _ *rankingv1.RankTopNRequest) (*rankingv1.RankTopNResponse, error) {
	return &rankingv1.RankTopNResponse{}, r.svc.RankTopN(ctx)
}

func (r *RankingServiceServer) TopN(ctx context.Context, _ *rankingv1.TopNRequest) (*rankingv1.TopNResponse, error) {
	domainAtcls, err := r.svc.TopN(ctx)
	if err != nil {
		return &rankingv1.TopNResponse{}, err
	}
	res := make([]*rankingv1.Article, 0, len(domainAtcls))
	for _, atcl := range domainAtcls {
		res = append(res, convertToV(atcl))
	}
	return &rankingv1.TopNResponse{
		Articles: res,
	}, err
}

func convertToV(domainArticle domain.Article) *rankingv1.Article {
	return &rankingv1.Article{
		Id:      domainArticle.Id,
		Title:   domainArticle.Title,
		Status:  int32(domainArticle.Status),
		Content: domainArticle.Content,
		Author: &rankingv1.Author{
			Id:   domainArticle.Author.Id,
			Name: domainArticle.Author.Name,
		},
		Ctime: timestamppb.New(domainArticle.Ctime),
		Utime: timestamppb.New(domainArticle.Utime),
	}
}
