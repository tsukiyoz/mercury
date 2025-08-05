package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/tsukiyo/mercury/pkg/grpcx/balancer/wrr"

	pb "github.com/tsukiyo/mercury/pkg/grpcx/examples/features/proto/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

const (
	exampleScheme      = "example"
	exampleServiceName = "lb.example.grpc"
)

var addrs = []string{"localhost:50051", "localhost:50052", "localhost:50053"}

func callUnaryEcho(ctx context.Context, c pb.EchoClient, msg string) {
	resp, err := c.UnaryEcho(ctx, &pb.EchoRequest{Message: msg})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	fmt.Println(resp)
}

func makeRPCs(ctx context.Context, cc *grpc.ClientConn, n int) {
	client := pb.NewEchoClient(cc)
	for range n {
		callUnaryEcho(ctx, client, "this is examples/load_balancing")
	}
}

func main() {
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		pickfirstConn, err := grpc.NewClient(
			fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer pickfirstConn.Close()

		fmt.Println("--- calling helloworld.Greeter/SayHello with pick_first ---")
		makeRPCs(ctx, pickfirstConn, 15)
		fmt.Println()
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		roundrobinConn, err := grpc.NewClient(
			fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName),
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer roundrobinConn.Close()

		fmt.Println("--- calling helloworld.Greeter/SayHello with round_robin ---")
		makeRPCs(ctx, roundrobinConn, 15)
		fmt.Println()
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		wrrConn, err := grpc.NewClient(
			fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName),
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"custom-wrr":{}}]}`),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer wrrConn.Close()

		fmt.Println("--- calling helloworld.Greeter/SayHello with wrr ---")
		makeRPCs(ctx, wrrConn, 15)
		fmt.Println()
	}
}

type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	r := &exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			exampleServiceName: addrs,
		},
	}
	r.start()
	return r, nil
}
func (*exampleResolverBuilder) Scheme() string { return exampleScheme }

type exampleResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *exampleResolver) start() {
	fmt.Println("start example resolver...")
	addrStrs := r.addrsStore[r.target.Endpoint()]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{
			Addr:       s,
			Attributes: attributes.New("weight", (float64(i)+1)*10),
		}
	}
	fmt.Println("resolver addrs:", addrs)
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
func (*exampleResolver) ResolveNow(resolver.ResolveNowOptions) {}
func (*exampleResolver) Close()                                {}

func init() {
	resolver.Register(&exampleResolverBuilder{})
}
