package grpcx

import (
	"context"
	"fmt"
	"github.com/tsukaychan/mercury/pkg/logger"
	"github.com/tsukaychan/mercury/pkg/netx"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"net"
	"strconv"
	"time"

	etcdv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	Name      string
	Port      int
	EtcdAddrs []string
	l         logger.Logger
	cancel    func()
	key       string
	em        endpoints.Manager
	client    *etcdv3.Client
}

func NewServer(srv *grpc.Server, name string, port int, etcdAddrs []string, l logger.Logger) *Server {
	return &Server{
		Server:    srv,
		Name:      name,
		Port:      port,
		EtcdAddrs: etcdAddrs,
		l:         l,
	}
}

func (srv *Server) Serve() error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(srv.Port))
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	srv.cancel = cancel
	err = srv.register(ctx)
	if err != nil {
		panic(err)
	}
	srv.l.Info("grpc server running", logger.String("name", srv.Name), logger.Int64("port", int64(srv.Port)))
	return srv.Server.Serve(lis)
}

func (srv *Server) register(ctx context.Context) error {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: srv.EtcdAddrs,
	})
	if err != nil {
		return err
	}
	srv.client = cli

	name := fmt.Sprintf("service/%s", srv.Name)
	addr := fmt.Sprintf("%s:%d", netx.GetOutboundIP(), srv.Port)
	srv.key = fmt.Sprintf("%s/%s", name, addr)

	em, err := endpoints.NewManager(cli, name)
	if err != nil {
		return err
	}

	// get keep alive lease
	var ttl int64 = 15
	lease, err := cli.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	err = em.AddEndpoint(ctx, srv.key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	// keep alive
	kaCh, err := cli.KeepAlive(ctx, lease.ID)
	if err != nil {
		srv.l.Error("keep alive failed", logger.Error(err))
	}
	go func() {
		for ka := range kaCh {
			srv.l.Debug("keep alive received", logger.String("message", ka.String()))
		}
	}()

	return nil
}

func (srv *Server) Close() error {
	if srv.cancel != nil {
		srv.cancel()
	}
	if srv.em != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := srv.em.DeleteEndpoint(ctx, srv.key)
		cancel()
		if err != nil {
			srv.l.Error("delete endpoint failed", logger.Error(err))
			return err
		}
	}
	if srv.client != nil {
		err := srv.client.Close()
		if err != nil {
			srv.l.Error("close etcd client failed", logger.Error(err))
			return err
		}
	}
	srv.Server.GracefulStop()
	return nil
}
