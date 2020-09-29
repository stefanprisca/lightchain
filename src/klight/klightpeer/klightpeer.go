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
	"log"
	"net"
	"os"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	lpack "github.com/stefanprisca/lightchain/src/lightpeer"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type klightpeer struct {
	*lpack.Lightpeer
}

func (klp *klightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	return klp.Lightpeer.Persist(ctx, tReq)
}

func (klp *klightpeer) Query(qReq *pb.EmptyQueryRequest, stream pb.Lightpeer_QueryServer) error {
	return klp.Lightpeer.Query(qReq, stream)
}

func (klp *klightpeer) JoinNetwork(ctx context.Context, joinReq *pb.JoinRequest) (*pb.JoinResponse, error) {
	return klp.Lightpeer.JoinNetwork(ctx, joinReq)
}

func (klp *klightpeer) ConnectNewPeer(cReq *pb.ConnectRequest, stream pb.Lightpeer_ConnectNewPeerServer) error {
	return klp.Lightpeer.ConnectNewPeer(cReq, stream)

}

func (klp *klightpeer) NotifyNewBlock(ctx context.Context, newBlock *pb.Lightblock) (*pb.NewBlockResponse, error) {
	return klp.Lightpeer.NotifyNewBlock(ctx, newBlock)
}

func (klp *klightpeer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return klp.Lightpeer.Check(ctx, in)
}

// GetState returns the current peer state
func (klp *klightpeer) GetState() pb.Lightblock {
	return klp.Lightpeer.GetState()
}

func startFileListener(statePath string) {
	stateFile, err := os.OpenFile(statePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.FileListener(stateFile)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := stateFile.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

}
