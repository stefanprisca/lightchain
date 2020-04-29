package main

import (
	"context"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
)

type lightpeer struct {
	pb.LightpeerServer
}

// Connect will connect to the network specified in the ConnectRequest
func (s *lightpeer) Connect(ctx context.Context, cReq *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	return &pb.ConnectResponse{}, nil
}

// Persist creates a new state on the chain, and notifies the network about the new state
func (s *lightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	return &pb.PersistResponse{}, nil
}

func (s *lightpeer) Query(qReq *pb.EmptyQueryRequest, stream pb.Lightpeer_QueryServer) error {
	return nil
}

func (s *lightpeer) NotifyNewBlock(ctx context.Context, nbReq *pb.NewBlockRequest) (*pb.NewBlockResponse, error) {
	return &pb.NewBlockResponse{}, nil
}
