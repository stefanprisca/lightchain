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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	lpb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
)

func TestKLightPeerListensToFile(t *testing.T) {

	statePath := "testdata/listenertest"
	os.Create(statePath)

	mockLP := &mockLightpeer{}
	klp := klightpeer{LightpeerServer: mockLP, statePath: statePath}
	go klp.startFileListener()

	ioutil.WriteFile(statePath, []byte("foo"), 0644)

	<-time.After(time.Second)
}

func TestKLightPeerSendsFileModifications(t *testing.T) {

	statePath := "testdata/listenertest"
	os.Create(statePath)

	mockLP := &mockLightpeer{}

	klp := klightpeer{LightpeerServer: mockLP, statePath: statePath}
	go klp.startFileListener()

	expectedPayload := []byte("foo")
	ioutil.WriteFile(statePath, expectedPayload, 0644)

	<-time.After(time.Second)
	assert.Contains(t, mockLP.messages, expectedPayload)
}

func TestKLightPeerWritesNewFiles(t *testing.T) {

	statePath := "testdata/listenertest"
	os.Create(statePath)

	mockLP := &mockLightpeer{}
	klp := klightpeer{LightpeerServer: mockLP, statePath: statePath}
	go klp.startFileListener()

	expectedPayload := []byte("foo")
	lb := &pb.Lightblock{Type: pb.Lightblock_CLIENT, Payload: expectedPayload}
	_, err := klp.NotifyNewBlock(context.Background(), lb)

	actualPayload, err := ioutil.ReadFile(statePath)
	assert.Nil(t, err)
	assert.Equal(t, expectedPayload, actualPayload)
}

func TestKLightPeerErrorsWhenNewBlockErrors(t *testing.T) {

	statePath := "testdata/listenertest"
	os.Create(statePath)

	mockLP := &mockLightpeer{errorOnNewBlock: fmt.Errorf("error")}
	klp := klightpeer{LightpeerServer: mockLP, statePath: statePath}

	lb := &pb.Lightblock{Type: pb.Lightblock_CLIENT, Payload: []byte("foo")}
	_, err := klp.NotifyNewBlock(context.Background(), lb)

	assert.NotNil(t, err)

	expectedPayload := []byte{}
	actualPayload, err := ioutil.ReadFile(statePath)
	assert.Nil(t, err)
	assert.Equal(t, expectedPayload, actualPayload)
}

func TestKLightPeerIgnoresNetworkBlocks(t *testing.T) {

	statePath := "testdata/listenertest"
	os.Create(statePath)

	mockLP := &mockLightpeer{}
	klp := klightpeer{LightpeerServer: mockLP, statePath: statePath}

	lb := &pb.Lightblock{Type: pb.Lightblock_NETWORK, Payload: []byte("foo")}
	_, err := klp.NotifyNewBlock(context.Background(), lb)

	assert.Nil(t, err)
	expectedPayload := []byte{}
	actualPayload, err := ioutil.ReadFile(statePath)
	assert.Nil(t, err)
	assert.Equal(t, expectedPayload, actualPayload)
}

type mockLightpeer struct {
	lpb.UnimplementedLightpeerServer

	messages        [][]byte
	errorOnNewBlock error
}

func (mlp *mockLightpeer) Persist(ctx context.Context, tReq *pb.PersistRequest) (*pb.PersistResponse, error) {
	mlp.messages = append(mlp.messages, tReq.Payload)
	return nil, nil
}

func (mlp *mockLightpeer) NotifyNewBlock(ctx context.Context, newBlock *pb.Lightblock) (*pb.NewBlockResponse, error) {
	return nil, mlp.errorOnNewBlock
}
