// Copyright 2020 Stefan Prisca
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/google/uuid"
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/api/trace"
)

type lightpeer struct {
	pb.LightpeerServer
	tr          trace.Tracer
	storagePath string
	state       pb.Lightblock
	network     []pb.PeerInfo
	meta        pb.PeerInfo
}

// Persist creates a new state on the chain, and notifies the network about the new state
func (lp *lightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	ctxt, span := lp.tr.Start(ctx, "persist")
	defer span.End()

	//log.Printf("got new persist request %v \n", *tReq)
	span.AddEvent(ctxt, fmt.Sprintf("got new persist request %v ", *tReq))

	lightBlock := pb.Lightblock{
		ID:      uuid.New().String(),
		Payload: tReq.Payload,
		Type:    pb.Lightblock_CLIENT,
	}

	if lp.state.ID != "" {
		lightBlock.PrevID = lp.state.ID
	}

	err := lp.writeBlock(lightBlock)
	if err == nil {
		lp.state = lightBlock
	}

	return &pb.PersistResponse{}, err
}

func (lp *lightpeer) Query(qReq *pb.EmptyQueryRequest, stream pb.Lightpeer_QueryServer) error {

	ctxt, span := lp.tr.Start(stream.Context(), "query")
	defer span.End()

	blockChan := lp.readBlocks()
	for blockResp := range blockChan {
		if blockResp.err != nil {
			span.AddEvent(ctxt, fmt.Sprintf("failed to read block %v \n", blockResp.err))
			return blockResp.err
		}
		if blockResp.block.Type != pb.Lightblock_CLIENT {
			continue
		}
		stream.Send(&pb.QueryResponse{Payload: blockResp.block.Payload})
	}

	span.AddEvent(ctxt, fmt.Sprintf("finished sending blocks \n"))

	return nil
}

// JoinNetwork makes a ConnectNewPeer request on the address given, and updates the internal peer state to match the newly joined netwrok.
func (lp *lightpeer) JoinNetwork(ctx context.Context, joinReq *pb.JoinRequest) (*pb.JoinResponse, error) {
	conn, err := grpc.Dial(joinReq.Address, grpc.WithInsecure())
	if err != nil {
		return &pb.JoinResponse{}, fmt.Errorf("did not connect: %s", err)
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewLightpeerClient(conn)
	pi := &pb.PeerInfo{}
	*pi = lp.meta
	blockStream, err := client.ConnectNewPeer(ctx, &pb.ConnectRequest{Peer: pi})
	if err != nil {
		// span.add event
		return &pb.JoinResponse{}, fmt.Errorf("could not connect new peer: %v", err)
	}
	networkUpdated := false
	var state *pb.Lightblock = nil
	for {
		block, err := blockStream.Recv()
		if err == io.EOF {
			// span.add event
			break
		}
		if err != nil {
			// span.add event
			return &pb.JoinResponse{}, fmt.Errorf("%v.Join returned error: %v", client, err)
		}
		err = lp.writeBlock(*block)
		if err != nil {
			// span.add event
			return &pb.JoinResponse{}, fmt.Errorf("could not write block %v: %v", *block, err)
		}

		if state == nil {
			state = block
		}

		if block.Type == pb.Lightblock_NETWORK && !networkUpdated {
			network := []pb.PeerInfo{}
			err := json.Unmarshal(block.Payload, &network)
			if err != nil {
				return &pb.JoinResponse{},
					fmt.Errorf("could not unmarshal network block: %v", err)
			}

			lp.network = network
			networkUpdated = true
		}
	}
	lp.state = *state
	if !networkUpdated {
		return &pb.JoinResponse{},
			fmt.Errorf("no network update blocks found, network state might be invalid")
	}
	return &pb.JoinResponse{}, nil
}

// Connect accepts connection from other peers.
func (lp *lightpeer) ConnectNewPeer(cReq *pb.ConnectRequest, stream pb.Lightpeer_ConnectNewPeerServer) error {

	ctxt, span := lp.tr.Start(stream.Context(), "connect")
	defer span.End()

	lp.network = append(lp.network, *cReq.Peer)

	rawNetwork, err := json.Marshal(lp.network)
	if err != nil {
		return fmt.Errorf("could not marshal new network: %v", err)
	}
	lightBlock := pb.Lightblock{
		ID:      uuid.New().String(),
		Payload: rawNetwork,
		Type:    pb.Lightblock_NETWORK,
		PrevID:  lp.state.ID,
	}

	err = lp.writeBlock(lightBlock)
	if err != nil {
		return fmt.Errorf("could not persist new network: %v", err)
	}
	lp.state = lightBlock

	blockChan := lp.readBlocks()
	for blockResp := range blockChan {
		if blockResp.err != nil {
			span.AddEvent(ctxt, fmt.Sprintf("failed to read block: %v \n", blockResp.err))
			return blockResp.err
		}

		lb := &pb.Lightblock{}
		*lb = blockResp.block

		stream.Send(lb)
	}

	return nil
}

type blockResponse struct {
	block pb.Lightblock
	err   error
}

func (lp *lightpeer) readBlocks() <-chan blockResponse {
	outchan := make(chan blockResponse, 1)
	go func() {
		defer close(outchan)

		outchan <- blockResponse{lp.state, nil}
		for blockID := lp.state.PrevID; blockID != ""; {
			blockFilePath := path.Join(lp.storagePath, blockID)
			rawBlock, err := ioutil.ReadFile(blockFilePath)
			if err != nil {
				outchan <- blockResponse{pb.Lightblock{}, err}
				return
			}

			block := &pb.Lightblock{}
			err = json.Unmarshal(rawBlock, block)
			if err != nil {
				outchan <- blockResponse{pb.Lightblock{}, err}
				return
			}
			outchan <- blockResponse{*block, nil}
			blockID = block.PrevID
		}
	}()

	return outchan
}

func (lp *lightpeer) writeBlock(block pb.Lightblock) error {

	out, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to encode block: %v", err)
	}
	outPath := path.Join(lp.storagePath, block.ID)

	if err := ioutil.WriteFile(outPath, out, 0666); err != nil {
		fmt.Errorf("failed to write block: %v", err)
	}

	return nil
}

func (lp *lightpeer) NotifyNewBlock(ctx context.Context, nbReq *pb.NewBlockRequest) (*pb.NewBlockResponse, error) {
	return &pb.NewBlockResponse{}, nil
}
