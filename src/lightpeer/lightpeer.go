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

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/plugin/grpctrace"
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
	persistCtx, span := lp.tr.Start(ctx, fmt.Sprintf("@%s - persist", lp.meta.Address))
	defer span.End()

	//log.Printf("got new persist request %v \n", *tReq)
	span.AddEvent(persistCtx, fmt.Sprintf("got new persist request %v ", *tReq))

	lightBlock := pb.Lightblock{
		ID:      uuid.New().String(),
		Payload: tReq.Payload,
		Type:    pb.Lightblock_CLIENT,
	}

	if lp.state.ID != "" {
		lightBlock.PrevID = lp.state.ID
	}

	err := lp.writeBlock(lightBlock)
	if err != nil {
		return &pb.PersistResponse{}, err
	}

	lp.state = lightBlock
	err = lp.sendNewBlockNotifications(persistCtx, lightBlock)
	return &pb.PersistResponse{}, err
}

func (lp *lightpeer) Query(qReq *pb.EmptyQueryRequest, stream pb.Lightpeer_QueryServer) error {

	queryCtx, span := lp.tr.Start(stream.Context(), fmt.Sprintf("@%s - query", lp.meta.Address))
	defer span.End()

	blockChan := lp.readBlocks()
	for blockResp := range blockChan {
		if blockResp.err != nil {
			span.RecordError(queryCtx, fmt.Errorf("failed to read block %v", blockResp.err))
			return blockResp.err
		}
		if blockResp.block.Type != pb.Lightblock_CLIENT {
			continue
		}
		stream.Send(&pb.QueryResponse{Payload: blockResp.block.Payload})
	}

	span.AddEvent(queryCtx, fmt.Sprintf("finished sending blocks \n"))

	return nil
}

// JoinNetwork makes a ConnectNewPeer request on the address given, and updates the internal peer state to match the newly joined netwrok.
func (lp *lightpeer) JoinNetwork(ctx context.Context, joinReq *pb.JoinRequest) (*pb.JoinResponse, error) {

	joinCtx, span := lp.tr.Start(ctx, fmt.Sprintf("@%s - join %s", lp.meta.Address, joinReq.Address))
	defer span.End()

	conn, err := grpc.Dial(joinReq.Address, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(
			global.Tracer(fmt.Sprintf("client@%s", joinReq.Address)))),
		grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(
			global.Tracer(fmt.Sprintf("stream-client@%s", joinReq.Address)))))

	if err != nil {
		err = fmt.Errorf("failed to connect to grpc server: %v", err)
		span.RecordError(joinCtx, err)
		return &pb.JoinResponse{}, err
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewLightpeerClient(conn)
	pi := &pb.PeerInfo{}
	*pi = lp.meta
	blockStream, err := client.ConnectNewPeer(joinCtx, &pb.ConnectRequest{Peer: pi})
	if err != nil {
		err := fmt.Errorf("connect new peer request failed: %v", err)
		span.RecordError(joinCtx, err)
		return &pb.JoinResponse{}, err
	}

	networkUpdated := false
	var state *pb.Lightblock = nil
	for {
		block, err := blockStream.Recv()
		if err == io.EOF {
			span.AddEvent(joinCtx, fmt.Sprintf("finished receiving messages"))
			break
		}
		if err != nil {
			err = fmt.Errorf("error while receiving messages: %v", err)
			span.RecordError(joinCtx, err)
			return &pb.JoinResponse{}, err
		}
		err = lp.writeBlock(*block)
		if err != nil {
			err = fmt.Errorf("error while writing new block: %v", err)
			span.RecordError(joinCtx, err)
			return &pb.JoinResponse{}, err
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
	span.AddEvent(joinCtx, fmt.Sprintf("successfully joined the network"))
	return &pb.JoinResponse{}, nil
}

// Connect accepts connection from other peers.
func (lp *lightpeer) ConnectNewPeer(cReq *pb.ConnectRequest, stream pb.Lightpeer_ConnectNewPeerServer) error {

	connectCtx, span := lp.tr.Start(stream.Context(), fmt.Sprintf("@%s - connect %s", lp.meta.Address, cReq.Peer.Address))
	defer span.End()

	lp.network = append(lp.network, *cReq.Peer)

	rawNetwork, err := json.Marshal(lp.network)
	if err != nil {
		err = fmt.Errorf("could not marshal new network: %v", err)
		span.RecordError(connectCtx, err)
		return err
	}
	lightBlock := pb.Lightblock{
		ID:      uuid.New().String(),
		Payload: rawNetwork,
		Type:    pb.Lightblock_NETWORK,
		PrevID:  lp.state.ID,
	}

	err = lp.writeBlock(lightBlock)
	if err != nil {
		err = fmt.Errorf("could not write new network block: %v", err)
		span.RecordError(connectCtx, err)
		return err
	}
	lp.state = lightBlock

	blockChan := lp.readBlocks()
	for blockResp := range blockChan {
		if blockResp.err != nil {
			err = fmt.Errorf("failed to read block: %v", blockResp.err)
			span.RecordError(connectCtx, err)
			return err
		}

		lb := &pb.Lightblock{}
		*lb = blockResp.block

		stream.Send(lb)
	}

	err = lp.sendNewBlockNotifications(connectCtx, lightBlock)
	if err != nil {
		err = fmt.Errorf("could not send new block notifications: %v", err)
		span.RecordError(connectCtx, err)
		return err
	}
	span.AddEvent(connectCtx, fmt.Sprintf("successfully connected new peer"))
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

func (lp *lightpeer) sendNewBlockNotifications(ctx context.Context, block pb.Lightblock) error {
	for _, peer := range lp.network {
		if peer.Address == lp.meta.Address {
			continue
		}

		conn, err := grpc.Dial(peer.Address, grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(
				global.Tracer(fmt.Sprintf("client@%s", peer.Address)))),
			grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(
				global.Tracer(fmt.Sprintf("stream-client@%s", peer.Address)))))
		if err != nil {
			return fmt.Errorf("did not connect: %s", err)
		}

		client := pb.NewLightpeerClient(conn)

		newBlock := &pb.Lightblock{}
		*newBlock = block
		_, err = client.NotifyNewBlock(ctx, newBlock)
		conn.Close()

		if err != nil {
			return fmt.Errorf("could not notify new block for %v: %v", peer.Address, err)
		}
	}

	return nil
}

func (lp *lightpeer) NotifyNewBlock(ctx context.Context, newBlock *pb.Lightblock) (*pb.NewBlockResponse, error) {
	notifyNewBlockCtx, span := lp.tr.Start(ctx, fmt.Sprintf("@%s - notifyNewBlock", lp.meta.Address))
	defer span.End()

	err := lp.writeBlock(*newBlock)
	if err != nil {
		err = fmt.Errorf("could not persist new block: %v", err)
		span.RecordError(notifyNewBlockCtx, err)
		return &pb.NewBlockResponse{}, err
	}
	lp.state = *newBlock

	if newBlock.Type == pb.Lightblock_NETWORK {
		network := []pb.PeerInfo{}
		err := json.Unmarshal(newBlock.Payload, &network)
		if err != nil {
			err = fmt.Errorf("could not unmarshal network block: %v", err)
			span.RecordError(notifyNewBlockCtx, err)
			return &pb.NewBlockResponse{}, err
		}

		lp.network = network
	}

	span.AddEvent(notifyNewBlockCtx, fmt.Sprintf("successfully recorded new block"))

	return &pb.NewBlockResponse{}, nil
}
