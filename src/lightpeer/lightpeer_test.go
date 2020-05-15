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
	"testing"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"go.opentelemetry.io/otel/api/global"
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
	lp := &lightpeer{
		storagePath: "./testdata",
		tr:          global.Tracer("test"),
	}
	ctxt := context.Background()

	message := "Hello from the test side!"
	persistReq := &pb.PersistRequest{
		Payload: []byte(message),
	}
	_, err := lp.Persist(ctxt, persistReq)
	if err != nil {
		t.Fail()
	}

	queryStream := mockQueryStream{nil, []*pb.QueryResponse{}}
	lp.Query(&pb.EmptyQueryRequest{}, &queryStream)

	if len(queryStream.responses) == 0 {
		t.Fatalf("No responses read after persisting one")
	}

	rsp := queryStream.responses[0]
	actualMessage := string(rsp.Payload)
	if message != actualMessage {
		t.Fatalf("got the wrong message back")
	}
}

func TestPersistSavesStateChain(t *testing.T) {
	lp := &lightpeer{
		storagePath: "./testdata",
		tr:          global.Tracer("test"),
	}
	ctxt := context.Background()

	messages := []string{
		"Hello", "from", "the", "test", "side!",
	}

	for _, msg := range messages {
		persistReq := &pb.PersistRequest{
			Payload: []byte(msg),
		}
		_, err := lp.Persist(ctxt, persistReq)
		if err != nil {
			t.Fail()
		}
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
		t.Log(actualMsg)
		if expectedMsg != actualMsg {
			t.Fatalf("got the wrong message back")
		}

	}
}
