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
	"io/ioutil"
	"path"

	"github.com/google/uuid"
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"

	"go.opentelemetry.io/otel/api/trace"
)

type lightpeer struct {
	pb.LightpeerServer

	tr trace.Tracer

	storagePath string
	state       pb.Lightblock
}

// Persist creates a new state on the chain, and notifies the network about the new state
func (lp *lightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	ctxt, span := lp.tr.Start(ctx, "persist")
	defer span.End()

	//log.Printf("got new persist request %v \n", *tReq)
	span.AddEvent(ctxt, fmt.Sprintf("got new persist request %v ", *tReq))

	lightBlock := &pb.Lightblock{
		ID:      uuid.New().String(),
		Payload: tReq.Payload,
		Type:    pb.Lightblock_CLIENT,
	}
	if lp.state.ID != "" {
		lightBlock.PrevID = lp.state.ID
	}
	lp.state = *lightBlock

	// log.Printf("processing new block %v \n", *lightBlock)

	span.AddEvent(ctxt, fmt.Sprintf("processing new block %v \n", *lightBlock))

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

	ctxt, span := lp.tr.Start(stream.Context(), "query")
	defer span.End()

	blockChan := lp.readBlocks(ctxt)
	for blockResp := range blockChan {
		if blockResp.err != nil {
			span.AddEvent(ctxt, fmt.Sprintf("failed to process connect \n"))
			return blockResp.err
		}

		stream.Send(&pb.QueryResponse{Payload: blockResp.block.Payload})
	}

	span.AddEvent(ctxt, fmt.Sprintf("finished sending blocks \n"))

	return nil
}

// Connect accepts connection from other peers.
func (lp *lightpeer) Connect(cReq *pb.ConnectRequest, stream pb.Lightpeer_ConnectServer) error {

	ctxt, span := lp.tr.Start(stream.Context(), "connect")
	defer span.End()

	blockChan := lp.readBlocks(ctxt)
	for blockResp := range blockChan {
		if blockResp.err != nil {
			span.AddEvent(ctxt, fmt.Sprintf("failed to process connect \n"))
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

func (lp *lightpeer) readBlocks(outterContext context.Context) <-chan blockResponse {
	outchan := make(chan blockResponse, 1)
	go func() {
		ctxt, span := lp.tr.Start(outterContext, "reading blocks")
		defer span.End()
		defer close(outchan)

		outchan <- blockResponse{lp.state, nil}
		for blockID := lp.state.PrevID; blockID != ""; {
			blockFilePath := path.Join(lp.storagePath, blockID)
			rawBlock, err := ioutil.ReadFile(blockFilePath)
			if err != nil {
				span.AddEvent(ctxt, fmt.Sprintf("failed to read block %v \n", err))
				outchan <- blockResponse{pb.Lightblock{}, err}
				return
			}

			block := &pb.Lightblock{}
			err = json.Unmarshal(rawBlock, block)
			if err != nil {
				span.AddEvent(ctxt, fmt.Sprintf("failed to read block %v \n", err))
				outchan <- blockResponse{pb.Lightblock{}, err}
				return
			}
			outchan <- blockResponse{*block, nil}
			blockID = block.PrevID
		}
		span.AddEvent(ctxt, fmt.Sprintf("finished reading all the blocks \n"))
	}()

	return outchan
}

func (lp *lightpeer) NotifyNewBlock(ctx context.Context, nbReq *pb.NewBlockRequest) (*pb.NewBlockResponse, error) {
	return &pb.NewBlockResponse{}, nil
}
