package otel

import (
	"context"
	"net"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	itest "github.com/tsukaychan/mercury/pkg/grpcx/interceptor/otel/test"
	"google.golang.org/grpc"
)

type InterceptorTestSuite struct {
	suite.Suite
}

func TestInterceptor(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}

func (s *InterceptorTestSuite) SetupSuite() {
	// itest.InitZipkin()
	itest.InitJaeger()
}

func (s *InterceptorTestSuite) TestServer() {
	t := s.T()
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			NewTraceInterceptorBuilder("server-test", nil, nil).BuildUnaryServerInterceptor(),
		),
	)
	defer server.GracefulStop()
	userServer := &itest.UserService{}
	itest.RegisterUserServiceServer(server, userServer)
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	err = server.Serve(lis)
	require.NoError(t, err)
}

func (s *InterceptorTestSuite) TestClient() {
	t := s.T()
	conn, err := grpc.NewClient(":8090",
		grpc.WithChainUnaryInterceptor(NewTraceInterceptorBuilder("client-test", nil, nil).BuildUnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	cli := itest.NewUserServiceClient(conn)

	for i := 0; i < 1; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Hour)
		spanCtx, span := otel.GetTracerProvider().Tracer("github.com/tsukaychan/mercury/pkg/grpcx/interceptor/otel").Start(ctx, "client_getbyid")
		resp, err := cli.GetByID(spanCtx, &itest.GetByIDReq{Id: 123})
		cancel()
		time.Sleep(time.Millisecond * 20)
		span.End()
		require.NoError(t, err)
		t.Log(resp.User)
	}
	time.Sleep(1 * time.Second)
}
