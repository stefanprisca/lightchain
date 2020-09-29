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
	"io/ioutil"
	"log"
	"os"
	"time"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type klightpeer struct {
	pb.LightpeerServer
	health.Server

	statePath   string
	lastModTime time.Time
}

func (klp *klightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	return klp.LightpeerServer.Persist(ctx, tReq)
}

func (klp *klightpeer) Query(qReq *pb.EmptyQueryRequest, stream pb.Lightpeer_QueryServer) error {
	return klp.LightpeerServer.Query(qReq, stream)
}

func (klp *klightpeer) JoinNetwork(ctx context.Context, joinReq *pb.JoinRequest) (*pb.JoinResponse, error) {
	return klp.LightpeerServer.JoinNetwork(ctx, joinReq)
}

func (klp *klightpeer) ConnectNewPeer(cReq *pb.ConnectRequest, stream pb.Lightpeer_ConnectNewPeerServer) error {
	return klp.LightpeerServer.ConnectNewPeer(cReq, stream)

}

func (klp *klightpeer) NotifyNewBlock(ctx context.Context, newBlock *pb.Lightblock) (*pb.NewBlockResponse, error) {
	newBlockRsvp, err := klp.LightpeerServer.NotifyNewBlock(ctx, newBlock)
	if err != nil || newBlock.Type == pb.Lightblock_NETWORK {
		return newBlockRsvp, err
	}
	klp.updateStateFile(newBlock)
	return newBlockRsvp, err
}

func (klp *klightpeer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

func (klp *klightpeer) startFileListener() {
	for {
		log.Println("Checking for filechanges")
		// Wait for a connection.
		stateFile, err := os.OpenFile(klp.statePath, os.O_RDONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		stats, err := stateFile.Stat()
		if err != nil {
			log.Fatal(err)
		}

		err = stateFile.Close()
		if err != nil {
			log.Fatal(err)
		}

		if stats.ModTime().After(klp.lastModTime) {
			log.Println("File changed!")
			klp.lastModTime = stats.ModTime()

			ctx := context.Background()
			payload, err := ioutil.ReadFile(klp.statePath)
			if err != nil {
				log.Fatal(err)
			}

			klp.Persist(ctx, &pb.PersistRequest{Payload: payload})
		}

		<-time.After(100 * time.Millisecond)
	}
}

func (klp *klightpeer) updateStateFile(block *pb.Lightblock) error {
	klp.lastModTime = time.Now()
	return ioutil.WriteFile(klp.statePath, block.Payload, 0644)
}
