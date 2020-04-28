package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"
)

type server struct {
	pb.LightpeerServer
}

func (s *server) Connect(ctx context.Context, cReq *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	return &pb.ConnectResponse{}, nil
}

func (s *server) Execute(ctx context.Context, tReq *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	return &pb.TransactionResponse{}, nil
}

func (s *server) NotifyNewBlock(ctx context.Context, nbReq *pb.NewBlockRequest) (*pb.NewBlockResponse, error) {
	return &pb.NewBlockResponse{}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9081))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterLightpeerServer(grpcServer, &server{})

	grpcServer.Serve(lis)
}
