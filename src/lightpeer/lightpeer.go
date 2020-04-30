package main

import (
	"context"
	"encoding/json"
	"fmt"
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

// Persist creates a new state on the chain, and notifies the network about the new state
func (lp *lightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	log.Printf("got new persist request %v \n", *tReq)

	lightBlock := &pb.Lightblock{
		ID:      uuid.New().String(),
		Payload: tReq.Payload,
	}
	if lp.state.ID != "" {
		lightBlock.PrevID = lp.state.ID
	}
	lp.state = *lightBlock

	log.Printf("processing new block %v \n", *lightBlock)

	out, err := json.Marshal(lightBlock)
	if err != nil {
		return &pb.PersistResponse{}, fmt.Errorf("failed to encode lightblock: %v", err)
	}
	outPath := path.Join(lp.storagePath, lightBlock.ID)

	if err := ioutil.WriteFile(outPath, out, 0644); err != nil {
		return &pb.PersistResponse{}, fmt.Errorf("failed to write lightblock: %v", err)
	}
	return &pb.PersistResponse{
		Response: outPath,
	}, nil
}

func (lp *lightpeer) Query(qReq *pb.EmptyQueryRequest, stream pb.Lightpeer_QueryServer) error {
	log.Printf("received query request\n")

	stream.Send(&pb.QueryResponse{Payload: lp.state.Payload})
	log.Printf("responded with block %v \n", lp.state)

	for blockID := lp.state.PrevID; blockID != ""; {
		blockFilePath := path.Join(lp.storagePath, blockID)
		rawBlock, err := ioutil.ReadFile(blockFilePath)
		if err != nil {
			return err
		}

		block := &pb.Lightblock{}
		err = json.Unmarshal(rawBlock, block)
		if err != nil {
			return err
		}

		stream.Send(&pb.QueryResponse{Payload: block.Payload})
		log.Printf("responded with block %v \n", block)
		blockID = block.PrevID
	}
	return nil
}

// Connect accepts connection from other peers.
func (lp *lightpeer) Connect(ctx context.Context, cReq *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	return &pb.ConnectResponse{}, nil
}

func (lp *lightpeer) NotifyNewBlock(ctx context.Context, nbReq *pb.NewBlockRequest) (*pb.NewBlockResponse, error) {
	return &pb.NewBlockResponse{}, nil
}
