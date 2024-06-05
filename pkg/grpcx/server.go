package grpcx

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/lazywoo/mercury/pkg/logger"
	"github.com/lazywoo/mercury/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	Name        string
	Port        int
	EtcdAddrs   []string
	EtcdTTL     int64
	EtcdClient  *etcdv3.Client
	etcdKey     string
	etcdManager endpoints.Manager
	cancel      func()
	l           logger.Logger
}

func NewServer(srv *grpc.Server, name string, port int, etcdAddrs []string, ttl int64, l logger.Logger) *Server {
	return &Server{
		Server:    srv,
		Name:      name,
		Port:      port,
		EtcdAddrs: etcdAddrs,
		EtcdTTL:   ttl,
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
		Endpoints:   srv.EtcdAddrs,
		DialTimeout: 3 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		return err
	}

	srv.EtcdClient = cli

	name := fmt.Sprintf("service/%s", srv.Name)
	addr := fmt.Sprintf("%s:%d", netx.GetOutboundIP(), srv.Port)
	srv.etcdKey = fmt.Sprintf("%s/%s", name, addr)

	em, err := endpoints.NewManager(cli, name)
	if err != nil {
		return err
	}

	// get keep alive lease
	lease, err := cli.Grant(ctx, srv.EtcdTTL)
	if err != nil {
		return err
	}

	err = em.AddEndpoint(ctx, srv.etcdKey, endpoints.Endpoint{
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
	if srv.etcdManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := srv.etcdManager.DeleteEndpoint(ctx, srv.etcdKey)
		cancel()
		if err != nil {
			srv.l.Error("delete endpoint failed", logger.Error(err))
			return err
		}
	}
	if srv.EtcdClient != nil {
		err := srv.EtcdClient.Close()
		if err != nil {
			srv.l.Error("close etcd EtcdClient failed", logger.Error(err))
			return err
		}
	}
	srv.Server.GracefulStop()
	return nil
}
