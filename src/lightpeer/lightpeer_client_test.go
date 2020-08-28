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
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"github.com/stretchr/testify/require"
	grpctrace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc"
	"go.opentelemetry.io/otel/api/global"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestPeerServerPersistsMessages(t *testing.T) {
	msg := "Hello from the peer client!"
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestPeerServerPersistsMessages")
	defer tn.stop()

	tn.startLPServer(8090).
		persist(8090, msg).
		assertExpectedMessages(msg)
}

func TestConnectUpdatesMessages(t *testing.T) {

	msg := "Hello from the peer client!"
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestConnectUpdatesMessages")
	defer tn.stop()

	tn.startLPServer(8090).
		persist(8090, msg).
		startLPServer(8091).
		connect(8091, 8090).
		assertExpectedMessages(msg)
}

func TestConnectUpdatesTopology(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestConnectUpdatesTopology")
	defer tn.stop()

	tn.startLPServer(8090).
		startLPServer(8091).
		connect(8091, 8090).
		assertNetworkTopology(8090, 8091)
}

// Test fails because peers don't notify new blocks yet
func TestThreePeerNetworkUpdatesMessages(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestThreePeerNetworkUpdatesMessages")
	defer tn.stop()

	tn.startLPServer(8081).
		persist(8081, "8081").
		startLPServer(8082).
		connect(8082, 8081).
		persist(8082, "8082").
		persist(8081, "8081#2").
		persist(8082, "8082#2").
		assertExpectedMessages("8082#2", "8081#2", "8082", "8081").
		startLPServer(8083).
		connect(8083, 8082).
		persist(8083, "8083").
		persist(8083, "8083#2").
		persist(8082, "8082#3").
		persist(8081, "8081#3").
		assertExpectedMessages("8081#3", "8082#3", "8083#2", "8083", "8082#2", "8081#2", "8082", "8081")
}

// Test fails because peers don't notify new blocks yet
func TestThreePeerNetworkUpdatesTopology(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestThreePeerNetworkUpdatesTopology")
	defer tn.stop()

	tn.startLPServer(8081).
		startLPServer(8082).
		connect(8082, 8081).
		startLPServer(8083).
		connect(8083, 8082).
		assertNetworkTopology(8081, 8082, 8083)
}

// Test fails because peers don't notify new blocks yet
func TestNotifyInvalidBlockRefused(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestThreePeerNetworkUpdatesTopology")
	defer tn.stop()

	tn.startLPServer(8081).
		startLPServer(8082).
		connect(8082, 8081).
		startLPServer(8083).
		connect(8083, 8082).
		persist(8083, "8083").
		expectFailure().notifyNewBlock(8082, pb.Lightblock{Type: pb.Lightblock_CLIENT}).assertFailed().
		expectFailure().notifyNewBlock(8081, pb.Lightblock{Type: pb.Lightblock_CLIENT, PrevID: uuid.New().String()}).assertFailed().
		expectFailure().notifyNewBlock(8083, pb.Lightblock{PrevID: uuid.New().String()}).assertFailed().
		assertExpectedMessages("8083")
}

func TestInvalidBlockDoesNotUpdateState(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestThreePeerNetworkUpdatesTopology")
	defer tn.stop()

	newState := &pb.Lightblock{}
	tn.startLPServer(8081).
		startLPServer(8082).
		connect(8082, 8081).
		persist(8082, "8082").
		withNewState(8081, "newState", newState).
		expectFailure().persist(8082, "8082#2").assertFailed().
		assertExpectedMessagesFor(8081, "newState", "8082").
		assertExpectedMessagesFor(8082, "8082")
}

func TestNotifyRecoversAfterStateMismatch(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestThreePeerNetworkUpdatesTopology")
	defer tn.stop()

	newState := &pb.Lightblock{}

	tn.startLPServer(8081).
		startLPServer(8082).
		connect(8082, 8081).
		persist(8082, "8082").
		assertExpectedMessages("8082").
		withNewState(8081, "newState", newState).
		expectFailure().persist(8082, "8082#2").assertFailed().
		assertExpectedMessagesFor(8081, "newState", "8082").
		assertExpectedMessagesFor(8082, "8082").
		notifyNewBlock(8082, *newState).
		assertExpectedMessages("newState", "8082").
		persist(8082, "8082#2").
		assertExpectedMessages("8082#2", "newState", "8082")

}

func TestNetworkRecoversAfterPeerFailure(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestThreePeerNetworkUpdatesTopology")
	defer tn.stop()

	tn.startLPServer(8091)

	time.Sleep(time.Second)
	tn.startLPServer(8092).
		connect(8092, 8091).
		startLPServer(8093).
		connect(8093, 8092).
		assertNetworkTopology(8091, 8092, 8093).
		stopLPServer(8092).
		persist(8091, "8091")

	time.Sleep(time.Second)

	tn.assertNetworkTopology(8091, 8093).
		startLPServer(8092).
		connect(8092, 8091).
		assertNetworkTopology(8091, 8093, 8092).
		assertExpectedMessages("8091")
}

func TestPeerSelfRecovery(t *testing.T) {
	tn := newTestNetwork(t) //.withOTLP(OTLPAddress, "TestThreePeerNetworkUpdatesTopology")
	defer tn.stop()

	tn.startLPServer(8091)

	time.Sleep(time.Second)
	tn.startLPServer(8092).
		connect(8092, 8091).
		startLPServer(8093).
		connect(8093, 8092).
		assertNetworkTopology(8091, 8092, 8093).
		stopLPServer(8092).
		persist(8091, "8091")

	time.Sleep(time.Second)

	tn.assertNetworkTopology(8091, 8093).
		startLPServer(8092).
		// In self-recovery mode, the peer should be able to regain access to the network from
		// the information stored in its persistant storage. e.g. recreate the chain from the blocks it knows about
		// connect(8092, 8091).
		assertNetworkTopology(8091, 8093, 8092).
		assertExpectedMessages("8091")
}

type testNetwork struct {
	test            *testing.T
	clients         map[int]testClient
	otelFinalizer   func() error
	ignoreNextError bool
}

func newTestNetwork(test *testing.T) *testNetwork {
	return &testNetwork{
		test:          test,
		clients:       make(map[int]testClient),
		otelFinalizer: func() error { return nil },
	}
}

func (tn *testNetwork) withOTLP(otlpBackend, serviceName string) *testNetwork {
	otelFinalizer := initOtel(otlpBackend, serviceName)
	tn.otelFinalizer = otelFinalizer
	return tn
}

func (tn *testNetwork) startLPServer(port int) *testNetwork {
	tc, err := startLPTestServer(port)
	tn.handleError("%v", err)
	tn.clients[port] = tc
	// sleep a bit to give the gRPC server a chance to start
	//time.Sleep(time.Second)

	return tn
}

func (tn *testNetwork) stopLPServer(port int) *testNetwork {

	tc := tn.clients[port]
	require.NoError(tn.test, tc.stop())
	delete(tn.clients, port)
	return tn
}

func (tn *testNetwork) persist(port int, messages ...string) *testNetwork {
	tc := tn.clients[port]
	ctx := getClientContext(tc)
	traceID := fmt.Sprintf("persist@client%d", port)
	persistCtx, span := global.Tracer(traceID).Start(ctx, traceID)
	defer span.End()

	for _, msg := range messages {
		persistReq := &pb.PersistRequest{
			Payload: []byte(msg),
		}
		_, err := tc.client.Persist(persistCtx, persistReq)
		tn.handleError("%v", err)
	}

	return tn
}
func (tn *testNetwork) connect(port, toPort int) *testNetwork {
	tc := tn.clients[port]
	ctx := getClientContext(tc)
	traceID := fmt.Sprintf("join@client%d", port)
	connectCtx, span := global.Tracer(traceID).Start(ctx, traceID)
	defer span.End()

	from := tc
	to := tn.clients[toPort]

	joinReq := &pb.JoinRequest{
		Address: to.lp.meta.Address,
	}
	_, err := from.client.JoinNetwork(connectCtx, joinReq)
	tn.handleError("%v", err)

	return tn
}

func (tn *testNetwork) notifyNewBlock(port int, block pb.Lightblock) *testNetwork {
	tc := tn.clients[port]
	ctx := getClientContext(tc)
	traceID := fmt.Sprintf("notifyNewBlock@client%d", port)
	notifyCtx, span := global.Tracer(traceID).Start(ctx, traceID)
	defer span.End()

	nb := &pb.Lightblock{}
	*nb = block

	_, err := tc.client.NotifyNewBlock(notifyCtx, nb)
	tn.handleError("notify new block returned with error: %v", err)

	return tn
}

func (tn *testNetwork) withNewState(port int, msg string, outState *pb.Lightblock) *testNetwork {
	tc := tn.clients[port]
	ctx := getClientContext(tc)
	traceID := fmt.Sprintf("changeState@client%d", port)
	_, span := global.Tracer(traceID).Start(ctx, traceID)
	defer span.End()

	newState := pb.Lightblock{
		ID:      uuid.New().String(),
		PrevID:  tc.lp.state.ID,
		Payload: []byte(msg),
		Type:    pb.Lightblock_CLIENT,
	}

	tn.notifyNewBlock(port, newState)
	*outState = newState

	return tn
}

func (tn *testNetwork) handleError(formatMsg string, err error) {
	if err == nil {
		return
	}

	if tn.ignoreNextError {
		tn.ignoreNextError = false
		return
	}

	tn.test.Fatalf(fmt.Sprintf(formatMsg, err))
}

func (tn *testNetwork) expectFailure() *testNetwork {
	tn.ignoreNextError = true
	return tn
}

func (tn *testNetwork) assertFailed() *testNetwork {
	if tn.ignoreNextError {
		tn.test.Fatalf("expected previous operation to fail")
	}
	return tn
}

func (tn *testNetwork) assertExpectedMessages(messages ...string) *testNetwork {
	for _, tc := range tn.clients {
		err := assertExpectedMessages(messages, tc)
		if err != nil {
			tn.test.Fatal(err)
		}
	}

	return tn
}

func (tn *testNetwork) assertExpectedMessagesFor(port int, messages ...string) *testNetwork {
	tc := tn.clients[port]
	err := assertExpectedMessages(messages, tc)
	if err != nil {
		tn.test.Fatal(err)
	}

	return tn
}

func (tn *testNetwork) assertNetworkTopology(orderedPorts ...int) *testNetwork {
	expectedPIs := []pb.PeerInfo{}
	for _, p := range orderedPorts {
		expectedPIs = append(expectedPIs, tn.clients[p].lp.meta)
	}

	for _, tc := range tn.clients {
		err := assertExpectedNetwork(expectedPIs, tc)
		if err != nil {
			tn.test.Fatal(err)
		}
	}

	return tn
}

func (tn *testNetwork) stop() {
	errors := []error{}
	for _, tc := range tn.clients {
		errors = append(errors, tc.stop())
		// sleep a bit to give the gRPC server a chance to close down
		// time.Sleep(time.Second)
	}

	errors = append(errors, tn.otelFinalizer())

	fail := false
	for _, err := range errors {
		if err != nil {
			tn.test.Error(err)
			fail = true
		}
	}
	if fail {
		tn.test.FailNow()
	}
}

type testClient struct {
	client pb.LightpeerClient
	lp     *lightpeer
	stop   func() error
}

func startLPTestServer(port int) (testClient, error) {

	blockRepoPath := fmt.Sprintf("./testdata/%d", port)
	os.MkdirAll(blockRepoPath, 0777)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return testClient{}, fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer, lp, nhc := newLPGrpcServer("", port, blockRepoPath)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	var conn *grpc.ClientConn
	clientTr := global.Tracer(fmt.Sprintf("client@%s", lp.meta.Address))
	conn, err = grpc.Dial(lp.meta.Address, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(clientTr)),
		grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(clientTr)))
	if err != nil {
		return testClient{}, fmt.Errorf("did not connect: %s", err)
	}
	// defer func() { _ = conn.Close() }()

	client := pb.NewLightpeerClient(conn)
	return testClient{
		client: client,
		lp:     lp,
		stop: func() error {
			clientError := conn.Close()
			grpcServer.Stop()
			nhc.stopPeerHealthCheck()
			return clientError
		}}, nil
}

func assertExpectedNetwork(expectedNetwork []pb.PeerInfo, tc testClient) error {
	pNetwork := tc.lp.network

	if len(pNetwork) != len(expectedNetwork) {
		return fmt.Errorf("expected %v peers, %v has %v",
			expectedNetwork, tc.lp.meta, pNetwork)
	}

	for i := 0; i < len(expectedNetwork); i++ {
		if pNetwork[i].Address != expectedNetwork[i].Address {
			return fmt.Errorf("%v has unexpected peer on position %d: expected %v, actual %v",
				tc.lp.meta, i, expectedNetwork[i].Address, pNetwork[i].Address)
		}
	}
	return nil
}

func assertExpectedMessages(expectedMessages []string, tc testClient) error {

	nOfMessages := len(expectedMessages)

	ctx := getClientContext(tc)
	traceID := fmt.Sprintf("query@client%s", tc.lp.meta.Address)
	queryCtx, span := global.Tracer(traceID).Start(ctx, traceID)
	defer span.End()

	queryClient, err := tc.client.Query(queryCtx, &pb.EmptyQueryRequest{})
	if err != nil {
		return err
	}

	for i := 0; i < nOfMessages; i++ {
		rsp, err := queryClient.Recv()
		if err == io.EOF {
			return fmt.Errorf("%v: unexpected end of query stream", tc.lp.meta)
		}
		if err != nil {
			return fmt.Errorf("%v.Query returned error: %v", tc.lp.meta, err)
		}

		expectedMessage := expectedMessages[i]
		actualMessage := string(rsp.Payload)
		if expectedMessage != actualMessage {
			return fmt.Errorf("got the wrong message for peer %v: expected %s, actual %s",
				tc.lp.meta, expectedMessage, actualMessage)
		}
	}
	return nil
}

func getClientContext(tc testClient) context.Context {
	address := tc.lp.meta.Address
	clientID := fmt.Sprintf("test-client@%s", address)
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", clientID,
	)

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return ctx
}
