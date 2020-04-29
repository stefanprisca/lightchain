package main

import (
	"context"
	"testing"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
)

func TestPersist(t *testing.T) {
	lp := &lightpeer{}

	ctxt := context.Background()
	t.Log("Created the context")
	persistReq := &pb.PersistRequest{}
	_, err := lp.Persist(ctxt, persistReq)

	if err != nil {
		t.Fail()
	}
}
