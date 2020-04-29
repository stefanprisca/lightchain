package main

import (
	"context"
	"testing"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"
)

func TestPersist(t *testing.T) {
	lp := &lightpeer{
		storagePath: "./testdata",
	}

	ctxt := context.Background()

	persistReq := &pb.PersistRequest{}
	_, err := lp.Persist(ctxt, persistReq)

	if err != nil {
		t.Fail()
	}
}

type mockBlockStream struct {
	grpc.ServerStream
	blocks []*pb.Lightblock
}

func (x *mockBlockStream) Send(m *pb.Lightblock) error {
	x.blocks = append(x.blocks, m)
	return nil
}

func TestPersistSavesState(t *testing.T) {
	lp := &lightpeer{
		storagePath: "./testdata",
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

	blockStream := mockBlockStream{nil, []*pb.Lightblock{}}
	lp.Query(&pb.EmptyQueryRequest{}, &blockStream)

	if len(blockStream.blocks) == 0 {
		t.Fatalf("No blocks read after persisting one")
	}

	firstBlock := blockStream.blocks[0]
	actualMessage := string(firstBlock.Payload)
	if message != actualMessage {
		t.Fatalf("got the wrong message back")
	}
}
