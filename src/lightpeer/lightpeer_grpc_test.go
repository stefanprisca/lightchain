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
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"go.opentelemetry.io/otel/api/global"
	"google.golang.org/grpc"
)

func TestPeerServerPersistsMessages(t *testing.T) {

	address := fmt.Sprintf(":%d", 8090)
	peerClient, terminator := startLPServer(address)
	defer terminator()
	msg := "Hello from the peer client!"
	persistReq := &pb.PersistRequest{
		Payload: []byte(msg),
	}
	ctx := context.Background()
	_, err := peerClient.Persist(ctx, persistReq)
	if err != nil {
		t.Fatal(err)
	}

	queryClient, err := peerClient.Query(ctx, &pb.EmptyQueryRequest{})
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := queryClient.Recv()
	if err == io.EOF {
		t.Fatalf("unexpected end of query stream")
	}
	if err != nil {
		t.Fatalf("%v.Query returned error: %v", peerClient, err)
	}

	actualMessage := string(rsp.Payload)
	if msg != actualMessage {
		t.Fatalf("got the wrong message back")
	}
}

func TestPeersConnect(t *testing.T) {
	addressA := fmt.Sprintf(":%d", 8090)
	peerA, terminatorA := startLPServer(addressA)
	defer terminatorA()

	msg := "Hello from the peer client!"
	persistReq := &pb.PersistRequest{
		Payload: []byte(msg),
	}
	ctx := context.Background()
	_, err := peerA.Persist(ctx, persistReq)
	if err != nil {
		t.Fatal(err)
	}

	addressB := fmt.Sprintf(":%d", 8091)
	peerB, terminatorB := startLPServer(addressB)
	defer terminatorB()

	_, err = peerB.JoinNetwork(ctx, &pb.JoinRequest{Address: addressA})
	if err != nil {
		t.Fatal(err)
	}

	queryClient, err := peerB.Query(ctx, &pb.EmptyQueryRequest{})
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := queryClient.Recv()
	if err == io.EOF {
		t.Fatalf("unexpected end of query stream")
	}
	if err != nil {
		t.Fatalf("%v.Query returned error: %v", peerB, err)
	}

	actualMessage := string(rsp.Payload)
	if msg != actualMessage {
		t.Fatalf("got the wrong message back")
	}
}

func startLPServer(address string) (pb.LightpeerClient, func() error) {

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLightpeerServer(grpcServer, &lightpeer{
		tr:          global.Tracer("foo"),
		storagePath: "./testdata",
	})

	go func() {
		log.Println("Start serving gRPC connections...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	var conn *grpc.ClientConn
	conn, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	// defer func() { _ = conn.Close() }()

	client := pb.NewLightpeerClient(conn)
	return client, func() error {
		clientError := conn.Close()
		grpcServer.Stop()
		return clientError
	}
}
