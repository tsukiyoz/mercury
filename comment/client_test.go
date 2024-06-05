package main

import (
	"context"
	"testing"

	commentv1 "github.com/lazywoo/mercury/api/proto/gen/comment/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCClientCommentArticle(t *testing.T) {
	// 评论一篇文章
	c, err := grpc.NewClient("localhost:8091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := commentv1.NewCommentServiceClient(c)
	{
		resp, err := client.CreateComment(context.Background(), &commentv1.CreateCommentRequest{
			Comment: &commentv1.Comment{
				Uid:     1,
				Biz:     "article",
				BizId:   1,
				Content: "这篇文章写的真好啊!",
			},
		})
		require.NoError(t, err)
		t.Log(resp)
	}

	{
		resp, err := client.CreateComment(context.Background(), &commentv1.CreateCommentRequest{
			Comment: &commentv1.Comment{
				Uid:     1,
				Biz:     "article",
				BizId:   1,
				Content: "真心不错，好吃不贵!",
			},
		})
		require.NoError(t, err)
		t.Log(resp)
	}

	{
		resp, err := client.CreateComment(context.Background(), &commentv1.CreateCommentRequest{
			Comment: &commentv1.Comment{
				Uid:     2,
				Biz:     "article",
				BizId:   1,
				Content: "还可以!",
			},
		})
		require.NoError(t, err)
		t.Log(resp)
	}

	{
		resp, err := client.CreateComment(context.Background(), &commentv1.CreateCommentRequest{
			Comment: &commentv1.Comment{
				Uid:     3,
				Biz:     "article",
				BizId:   1,
				Content: "不错的!",
			},
		})
		require.NoError(t, err)
		t.Log(resp)
	}
}

func TestGRPCClientCommentReply(t *testing.T) {
	// 回复一篇评论
	c, err := grpc.NewClient("localhost:8091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := commentv1.NewCommentServiceClient(c)
	{
		resp, err := client.CreateComment(context.Background(), &commentv1.CreateCommentRequest{
			Comment: &commentv1.Comment{
				Uid:     2,
				Biz:     "article",
				BizId:   1,
				Content: "可恶的小日本!",
				RootComment: &commentv1.Comment{
					Id: 5,
				},
				ParentComment: &commentv1.Comment{
					Id: 5,
				},
			},
		})
		require.NoError(t, err)
		t.Log(resp)
	}

	{
		resp, err := client.CreateComment(context.Background(), &commentv1.CreateCommentRequest{
			Comment: &commentv1.Comment{
				Uid:     2,
				Biz:     "article",
				BizId:   1,
				Content: "难不成第三次世界大战要开始了!",
				RootComment: &commentv1.Comment{
					Id: 5,
				},
				ParentComment: &commentv1.Comment{
					Id: 5,
				},
			},
		})
		require.NoError(t, err)
		t.Log(resp)
	}

	{
		resp, err := client.CreateComment(context.Background(), &commentv1.CreateCommentRequest{
			Comment: &commentv1.Comment{
				Uid:     3,
				Biz:     "article",
				BizId:   1,
				Content: "经验+3!",
				RootComment: &commentv1.Comment{
					Id: 5,
				},
				ParentComment: &commentv1.Comment{
					Id: 5,
				},
			},
		})
		require.NoError(t, err)
		t.Log(resp)
	}
}

func TestGRPCClientGetFirstLevelComment(t *testing.T) {
	// 得到第一级评论
	c, err := grpc.NewClient("localhost:8091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := commentv1.NewCommentServiceClient(c)
	resp, err := client.GetCommentList(context.Background(), &commentv1.CommentListRequest{
		Biz:   "article",
		BizId: 1,
		Limit: 3,
	})
	t.Log(resp)
}

func TestGRPCClientDeleteComment(t *testing.T) {
	// 删除一篇评论，子评论也会随之被删除
	c, err := grpc.NewClient("localhost:8091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := commentv1.NewCommentServiceClient(c)
	resp, err := client.DeleteComment(context.Background(), &commentv1.DeleteCommentRequest{
		Id: 2,
	})
	t.Log(resp)
}

func TestGRPCClientGetReplies(t *testing.T) {
	// 得到子评论
	c, err := grpc.NewClient("localhost:8091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := commentv1.NewCommentServiceClient(c)
	resp, err := client.GetMoreReplies(context.Background(), &commentv1.GetMoreRepliesRequest{
		Rid:   5,
		MaxId: 13,
		Limit: 3,
	})
	t.Log(resp)
}
