package grpcx

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	Addr string
}

func (srv *Server) Serve() error {
	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		panic(err)
	}
	log.Println("server running at", srv.Addr)
	return srv.Server.Serve(lis)
}
