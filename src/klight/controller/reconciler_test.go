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
	"log"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lpb "github.com/stefanprisca/lightchain/src/api/lightpeer"
)

func TestReconcilerStacksIPs(t *testing.T) {
	rtc := newReconcilerTestCase(t)
	defer rtc.stop()

	rtc.addPod(":8081", "networkId").
		addPod(":8082", "networkId").
		addPod(":8083", "networkId4").
		addPod(":8084", "networkId2").
		assertReconcilerStacks(map[string][]string{
			"networkId":  {":8082", ":8081"},
			"networkId4": {":8083"},
			"networkId2": {":8084"},
		})
}

func TestReconcilerConnectsPods(t *testing.T) {
	rtc := newReconcilerTestCase(t)
	defer rtc.stop()

	rtc.addPod(":8081", "networkId").
		addPod(":8082", "networkId").
		assertPodsAreConnected(map[string][]string{
			"networkId": {":8082", ":8081"},
		})
}

type klightTestPod struct {
	lpb.UnimplementedLightpeerServer

	pod       v1.Pod
	lpMeta    lpb.PeerInfo
	lpNetwork []string
	stop      func()
}

func (ktp *klightTestPod) JoinNetwork(ctx context.Context, joinReq *lpb.JoinRequest) (*lpb.JoinResponse, error) {
	conn, err := grpc.Dial(joinReq.Address, grpc.WithInsecure())

	if err != nil {
		return &lpb.JoinResponse{}, err
	}
	defer func() { _ = conn.Close() }()

	client := lpb.NewLightpeerClient(conn)
	pi := &lpb.PeerInfo{}
	*pi = ktp.lpMeta
	_, err = client.ConnectNewPeer(context.Background(), &lpb.ConnectRequest{Peer: pi})
	if err != nil {
		return &lpb.JoinResponse{}, err
	}

	ktp.lpNetwork = append(ktp.lpNetwork, joinReq.Address)

	return &lpb.JoinResponse{}, nil
}

func (ktp *klightTestPod) ConnectNewPeer(cReq *lpb.ConnectRequest, stream lpb.Lightpeer_ConnectNewPeerServer) error {
	ktp.lpNetwork = append(ktp.lpNetwork, cReq.Peer.Address)
	return nil
}

func (ktp *klightTestPod) startGrpc() error {
	lis, err := net.Listen("tcp", ktp.lpMeta.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	lpb.RegisterLightpeerServer(grpcServer, ktp)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	ktp.stop = func() { grpcServer.Stop() }
	return nil
}

type reconcilerTestCase struct {
	nr       *networkReconciler
	t        *testing.T
	testPods map[string][]*klightTestPod
}

func newReconcilerTestCase(t *testing.T) reconcilerTestCase {
	return reconcilerTestCase{
		nr:       &networkReconciler{map[string]ipStack{}},
		t:        t,
		testPods: map[string][]*klightTestPod{},
	}
}

func (rtc reconcilerTestCase) addPod(ip, networkId string) reconcilerTestCase {
	pod := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{Labels: map[string]string{klightNetworkLabel: networkId}},
		Status:     v1.PodStatus{PodIP: ip},
	}

	klpTestPod := &klightTestPod{
		pod:       pod,
		lpMeta:    lpb.PeerInfo{Address: ip},
		lpNetwork: []string{},
	}

	err := klpTestPod.startGrpc()
	if err != nil {
		rtc.stop()
		rtc.t.Fatalf("filed to add pod: %v", err)
	}

	rtc.testPods[networkId] = append(rtc.testPods[networkId], klpTestPod)

	time.Sleep(time.Second)

	rtc.nr.reconcileLightNetwork(pod)
	return rtc
}

func (rtc reconcilerTestCase) stop() {
	for _, ktps := range rtc.testPods {
		for _, ktp := range ktps {
			ktp.stop()
		}
	}
}

func (rtc reconcilerTestCase) assertReconcilerStacks(expectedNetworks map[string][]string) {
	for netId, expectedStack := range expectedNetworks {
		actualStack := rtc.nr.stacks[netId].asList()

		if len(expectedStack) != len(actualStack) {
			rtc.t.Fatalf("expected stack for network < %s > does not match actual stack: \n expected: %v \n actual: %v ",
				netId, expectedStack, actualStack)
		}

		for i, ip := range expectedStack {
			if actualStack[i] != ip {
				rtc.t.Fatalf("expected stack for network < %s >  does not match actual stack: \n expected: %v \n actual: %v ",
					netId, expectedStack, actualStack)
			}
		}
	}
}

func (rtc reconcilerTestCase) assertPodsAreConnected(expectedNetworks map[string][]string) {
	// for _, expectedStack := range expectedNetworks {

	// }

}
