package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"path"

	"github.com/google/uuid"
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
)

type lightpeer struct {
	pb.LightpeerServer

	storagePath string
	state       pb.Lightblock
}

// Connect will connect to the network specified in the ConnectRequest
func (lp *lightpeer) Connect(ctx context.Context, cReq *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	return &pb.ConnectResponse{}, nil
}

// Persist creates a new state on the chain, and notifies the network about the new state
func (lp *lightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	lightBlock := &pb.Lightblock{
		ID:      uuid.New().String(),
		Payload: tReq.Payload,
	}
	lp.state = *lightBlock

	out, err := json.Marshal(lightBlock)
	if err != nil {
		log.Fatalln("Failed to encode lightblock:", err)
	}

	outPath := path.Join(lp.storagePath, lightBlock.ID)

	if err := ioutil.WriteFile(outPath, out, 0644); err != nil {
		log.Fatalln("Failed to write lightblock:", err)
	}
	return &pb.PersistResponse{
		Response: outPath,
	}, nil
}

func (lp *lightpeer) Query(qReq *pb.EmptyQueryRequest, stream pb.Lightpeer_QueryServer) error {
	stream.Send(&lp.state)
	return nil
}

func (lp *lightpeer) NotifyNewBlock(ctx context.Context, nbReq *pb.NewBlockRequest) (*pb.NewBlockResponse, error) {
	return &pb.NewBlockResponse{}, nil
}
