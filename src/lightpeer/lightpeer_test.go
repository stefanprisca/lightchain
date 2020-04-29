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

type mockQueryStream struct {
	grpc.ServerStream
	responses []*pb.QueryResponse
}

func (x *mockQueryStream) Send(m *pb.QueryResponse) error {
	x.responses = append(x.responses, m)
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
