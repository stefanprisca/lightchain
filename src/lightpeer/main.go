package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/stefanprisca/lightchain/src/api/lightpeer"
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"
)

type server struct {
	pb.LightpeerServer
}

func (s *server) Connect(ctx context.Context, cReq *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	return &pb.ConnectResponse{}, nil
}

func (s *server) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	return &pb.PersistResponse{}, nil
}

func (s *server) Query(qReq *pb.EmptyQueryRequest, stream lightpeer.Lightpeer_QueryServer) error {
	return nil
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
