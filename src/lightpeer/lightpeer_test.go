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
	"testing"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/stdout"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

func TestPersist(t *testing.T) {
	lp := &lightpeer{
		storagePath: "./testdata",
		tr:          global.Tracer("test"),
	}

	ctxt := context.Background()

	persistReq := &pb.PersistRequest{}
	_, err := lp.Persist(ctxt, persistReq)

	if err != nil {
		t.Fail()
	}
}

type mockQueryStream struct {
	grpc.ServerStream
	responses []*pb.QueryResponse
}

func (x *mockQueryStream) Send(m *pb.QueryResponse) error {
	x.responses = append(x.responses, m)
	return nil
}

func (x *mockQueryStream) Context() context.Context {
	return context.Background()
}

func TestPersistSavesState(t *testing.T) {
	messages := []string{
		"Hello",
	}
	lp, err := initPeerFromBlocks(messages, false)
	if err != nil {
		t.Fatal(err)
	}

	queryStream := mockQueryStream{nil, []*pb.QueryResponse{}}
	lp.Query(&pb.EmptyQueryRequest{}, &queryStream)

	if len(queryStream.responses) == 0 {
		t.Fatalf("No responses read after persisting one")
	}

	rsp := queryStream.responses[0]
	actualMessage := string(rsp.Payload)
	if messages[0] != actualMessage {
		t.Fatalf("got the wrong message back")
	}
}

func TestPersistSavesStateChain(t *testing.T) {
	messages := []string{
		"Hello", "from", "the", "test", "side!",
	}
	lp, err := initPeerFromBlocks(messages, false)
	if err != nil {
		t.Fatal(err)
	}

	queryStream := mockQueryStream{nil, []*pb.QueryResponse{}}
	lp.Query(&pb.EmptyQueryRequest{}, &queryStream)

	expectedLength := len(messages)
	if len(queryStream.responses) != expectedLength {
		t.Fatalf("not all messages retrived after persisting")
	}

	for i := 0; i < expectedLength; i++ {
		expectedMsg := messages[i]
		rsp := queryStream.responses[expectedLength-i-1]

		actualMsg := string(rsp.Payload)
		if expectedMsg != actualMsg {
			t.Fatalf("got the wrong message back")
		}

	}
}

type mockLBStream struct {
	grpc.ServerStream
	responses []*pb.Lightblock
}

func (x *mockLBStream) Send(m *pb.Lightblock) error {
	x.responses = append(x.responses, m)
	return nil
}

func (x *mockLBStream) Context() context.Context {
	return context.Background()
}

func TestConnectReturnsExistingBlocks(t *testing.T) {

	messages := []string{
		"Hello", "from", "the", "test", "side!",
	}

	lp, err := initPeerFromBlocks(messages, false)
	if err != nil {
		t.Fatal(err)
	}

	stream := mockLBStream{nil, []*pb.Lightblock{}}
	lp.Connect(&pb.ConnectRequest{}, &stream)

	expectedLength := len(messages)
	actualLength := len(stream.responses)
	if actualLength != expectedLength {
		t.Fatalf("not all messages returned during connect: expected %v, got %v",
			expectedLength, actualLength)
	}
	for i := 0; i < expectedLength; i++ {
		expectedMsg := messages[i]
		rsp := stream.responses[expectedLength-i-1]

		actualMsg := string(rsp.Payload)
		if expectedMsg != actualMsg {
			t.Fatalf("got the wrong message back: expected %s, got %s",
				expectedMsg, actualMsg)
		}
	}
}

func TestConnectReturnsNetworkTopology(t *testing.T) {
	lp, err := initPeerFromBlocks([]string{}, false)
	if err != nil {
		t.Fatal(err)
	}

	stream := mockLBStream{nil, []*pb.Lightblock{}}
	lp.Connect(&pb.ConnectRequest{}, &stream)

	expectedLength := 1
	if len(stream.responses) != expectedLength {
		t.Fatalf("connect did not return any messages")
	}

	netTop := stream.responses[0]
	if netTop.Type != pb.Lightblock_NETWORK {
		t.Fatalf("connect did not return network topology")
	}
}

func initTestOtel() {
	stdOutExp, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
	if err != nil {
		log.Fatal(err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(stdOutExp))
	if err != nil {
		log.Fatalf("error creating trace provider: %v\n", err)
	}

	global.SetTraceProvider(tp)
}

func initPeerFromBlocks(messages []string, verbose bool) (*lightpeer, error) {
	if verbose {
		initTestOtel()
	}
	lp := &lightpeer{
		storagePath: "./testdata",
		tr:          global.Tracer("test"),
	}
	ctxt := context.Background()

	for _, msg := range messages {
		persistReq := &pb.PersistRequest{
			Payload: []byte(msg),
		}
		_, err := lp.Persist(ctxt, persistReq)
		if err != nil {
			return nil, err
		}
	}
	return lp, nil
}
