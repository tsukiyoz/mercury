package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/lazywoo/mercury/internal/article/domain"

	"google.golang.org/grpc"

	articlev1 "github.com/lazywoo/mercury/api/gen/article/v1"
	"github.com/lazywoo/mercury/internal/article/service"
)

type ArticleServiceServer struct {
	articlev1.UnimplementedArticleServiceServer
	service service.ArticleService
}

func NewArticleServiceServer(svc service.ArticleService) *ArticleServiceServer {
	return &ArticleServiceServer{
		service: svc,
	}
}

func (a *ArticleServiceServer) Register(server grpc.ServiceRegistrar) {
	articlev1.RegisterArticleServiceServer(server, a)
}

func (a *ArticleServiceServer) Save(ctx context.Context, req *articlev1.SaveRequest) (*articlev1.SaveResponse, error) {
	id, err := a.service.Save(ctx, convertToDomain(req.Article))
	return &articlev1.SaveResponse{Id: id}, err
}

func (a *ArticleServiceServer) Publish(ctx context.Context, req *articlev1.PublishRequest) (*articlev1.PublishResponse, error) {
	id, err := a.service.Publish(ctx, convertToDomain(req.Article))
	return &articlev1.PublishResponse{Id: id}, err
}

func (a *ArticleServiceServer) Withdraw(ctx context.Context, req *articlev1.WithdrawRequest) (*articlev1.WithdrawResponse, error) {
	err := a.service.Withdraw(ctx, req.Id, req.Uid)
	return &articlev1.WithdrawResponse{}, err
}

func (a *ArticleServiceServer) List(ctx context.Context, req *articlev1.ListRequest) (*articlev1.ListResponse, error) {
	articles, err := a.service.List(ctx, req.GetAuthor(), int(req.GetOffset()), int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	res := make([]*articlev1.Article, 0, len(articles))
	for _, article := range articles {
		res = append(res, convertToV(article))
	}
	return &articlev1.ListResponse{Articles: res}, nil
}

func (a *ArticleServiceServer) GetById(ctx context.Context, req *articlev1.GetByIdRequest) (*articlev1.GetByIdResponse, error) {
	article, err := a.service.GetById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &articlev1.GetByIdResponse{Article: convertToV(article)}, nil
}

func (a *ArticleServiceServer) GetPublishedById(ctx context.Context, req *articlev1.GetPublishedByIdRequest) (*articlev1.GetPublishedByIdResponse, error) {
	article, err := a.service.GetPublishedById(ctx, req.GetId(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &articlev1.GetPublishedByIdResponse{Article: convertToV(article)}, nil
}

func (a *ArticleServiceServer) ListPub(ctx context.Context, req *articlev1.ListPubRequest) (*articlev1.ListPubResponse, error) {
	atcls, err := a.service.ListPub(ctx, req.GetStartTime().AsTime(), int(req.GetOffset()), int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	list := make([]*articlev1.Article, 0, len(atcls))
	for _, atcl := range atcls {
		list = append(list, convertToV(atcl))
	}
	return &articlev1.ListPubResponse{Articles: list}, nil
}

func convertToV(domainArticle domain.Article) *articlev1.Article {
	return &articlev1.Article{
		Id:      domainArticle.Id,
		Title:   domainArticle.Title,
		Status:  int32(domainArticle.Status),
		Content: domainArticle.Content,
		Author: &articlev1.Author{
			Id:   domainArticle.Author.Id,
			Name: domainArticle.Author.Name,
		},
		Ctime:    timestamppb.New(domainArticle.Ctime),
		Utime:    timestamppb.New(domainArticle.Utime),
		Abstract: domainArticle.Abstract(),
	}
}

func convertToDomain(vArticle *articlev1.Article) domain.Article {
	return domain.Article{
		Id:      vArticle.GetId(),
		Title:   vArticle.GetTitle(),
		Content: vArticle.GetContent(),
		Author: domain.Author{
			Id:   vArticle.GetAuthor().GetId(),
			Name: vArticle.GetAuthor().GetName(),
		},
		Status: domain.ArticleStatus(vArticle.GetStatus()),
		Ctime:  vArticle.GetCtime().AsTime(),
		Utime:  vArticle.GetUtime().AsTime(),
	}
}
