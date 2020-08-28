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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lpb "github.com/stefanprisca/lightchain/src/api/lightpeer"
)

func TestReconcilerBuildsNetworks(t *testing.T) {
	rtc := newReconcilerTestCase(t)
	defer rtc.stop()

	rtc = rtc.addPod(8081, "networkId").
		addPod(8082, "networkId").
		addPod(8083, "networkId4").
		addPod(8084, "networkId2")

	expectedStacks := map[string][]int32{
		"networkId":  {8082, 8081},
		"networkId4": {8083},
		"networkId2": {8084},
	}

	rtc.assertReconcilerStacks(expectedStacks)

	time.Sleep(2 * time.Second)
	expectedNetworks := map[int32][]int32{
		8082: {8082, 8081},
		8081: {8082, 8081},
		8083: {8083},
		8084: {8084},
	}

	rtc.assertPodConnections(expectedNetworks)
}

func TestReconcilerMaintainsStacksAfterPodFailure(t *testing.T) {
	rtc := newReconcilerTestCase(t)
	defer rtc.stop()

	rtc = rtc.addPod(8081, "networkId").
		addPod(8082, "networkId").
		stopPod(8082).
		addPod(8083, "networkId")

	expectedStacks := map[string][]int32{
		"networkId": {8083, 8081},
	}

	rtc.assertReconcilerStacks(expectedStacks)

	time.Sleep(2 * time.Second)
	expectedNetworks := map[int32][]int32{
		8081: {8083, 8081},
		8083: {8083, 8081},
	}

	rtc.assertPodConnections(expectedNetworks)
}

func TestReconcilerMaintainsStacksAfterPodRejoins(t *testing.T) {
	rtc := newReconcilerTestCase(t)
	defer rtc.stop()

	rtc = rtc.addPod(8081, "networkId").
		addPod(8082, "networkId").
		stopPod(8082).
		addPod(8083, "networkId").
		addPod(8082, "networkId")

	expectedStacks := map[string][]int32{
		"networkId": {8082, 8083, 8081},
	}

	rtc.assertReconcilerStacks(expectedStacks)

	time.Sleep(2 * time.Second)
	expectedNetworks := map[int32][]int32{
		8081: {8083, 8081},
		8083: {8082, 8083, 8081},
		8082: {8083, 8082},
	}

	rtc.assertPodConnections(expectedNetworks)
}

func TestReconcilerMaintainsStacksAfterMultiplePodRestarts(t *testing.T) {
	rtc := newReconcilerTestCase(t)
	defer rtc.stop()

	rtc = rtc.addPod(8081, "networkId").
		addPod(8082, "networkId").
		stopPod(8082).
		addPod(8083, "networkId").
		addPod(8084, "networkId2").
		addPod(8085, "networkId2").
		addPod(8082, "networkId").
		stopPod(8084).
		addPod(8084, "networkId2").
		stopPod(8084).
		addPod(8084, "networkId2").
		stopPod(8084).
		addPod(8084, "networkId2")

	expectedStacks := map[string][]int32{
		"networkId":  {8082, 8083, 8081},
		"networkId2": {8084, 8085, 8084},
	}

	rtc.assertReconcilerStacks(expectedStacks)

	time.Sleep(2 * time.Second)
	expectedNetworks := map[int32][]int32{
		8081: {8083, 8082, 8081},
		8083: {8082, 8083, 8081},
		8082: {8083, 8082},
		8084: {8084, 8085},
		8085: {8084, 8085},
	}

	rtc.assertPodConnections(expectedNetworks)
}

type klightTestPod struct {
	lpb.UnimplementedLightpeerServer

	pod       v1.Pod
	lpMeta    lpb.PeerInfo
	lpNetwork []string
	server    *grpc.Server

	rtc *reconcilerTestCase
}

func (ktp *klightTestPod) JoinNetwork(ctx context.Context, joinReq *lpb.JoinRequest) (*lpb.JoinResponse, error) {

	log.Println("@Join: received req to join network", joinReq.Address)

	i, _ := strconv.ParseInt(strings.Split(joinReq.Address, ":")[1], 0, 32)
	otherPort := int32(i)
	connectTo, ok := ktp.rtc.testPods[otherPort]
	if !ok {
		return &lpb.JoinResponse{}, fmt.Errorf("peer unavailable")
	}

	connectTo.lpNetwork = append(connectTo.lpNetwork, ktp.lpMeta.Address)

	ktp.lpNetwork = append(ktp.lpNetwork, joinReq.Address)

	return &lpb.JoinResponse{}, nil
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

	ktp.server = grpcServer
	return nil
}

type reconcilerTestCase struct {
	nr       *networkReconciler
	t        *testing.T
	testPods map[int32]*klightTestPod
}

func newReconcilerTestCase(t *testing.T) *reconcilerTestCase {
	return &reconcilerTestCase{
		nr:       &networkReconciler{map[string]addressStack{}},
		t:        t,
		testPods: map[int32]*klightTestPod{},
	}
}

func (rtc *reconcilerTestCase) addPod(port int32, networkId string) *reconcilerTestCase {
	pod := v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{Labels: map[string]string{klightNetworkLabel: networkId}},
		Spec: v1.PodSpec{Containers: []v1.Container{v1.Container{
			Ports: []v1.ContainerPort{
				v1.ContainerPort{Name: klightPodPort, ContainerPort: port}}}}},
	}

	podAddress := fmt.Sprintf(":%d", port)
	klpTestPod := &klightTestPod{
		pod:       pod,
		lpMeta:    lpb.PeerInfo{Address: podAddress},
		lpNetwork: []string{podAddress},
		rtc:       rtc,
	}

	err := klpTestPod.startGrpc()
	rtc.handleError(err)

	rtc.testPods[port] = klpTestPod

	err = rtc.nr.reconcileLightNetwork(pod)
	rtc.handleError(err)

	return rtc
}

func (rtc *reconcilerTestCase) stopPod(port int32) *reconcilerTestCase {
	podToRemove := rtc.testPods[port]
	podToRemove.server.Stop()
	delete(rtc.testPods, port)

	time.Sleep(2 * time.Second)
	return rtc
}

func (rtc *reconcilerTestCase) handleError(err error) {
	if err != nil {
		rtc.stop()
		rtc.t.Fatalf("%v", err)
	}
}

func (rtc *reconcilerTestCase) stop() {
	for _, ktp := range rtc.testPods {
		ktp.server.Stop()
	}
}

func (rtc *reconcilerTestCase) assertReconcilerStacks(expectedNetworks map[string][]int32) *reconcilerTestCase {
	for netID, expectedStack := range expectedNetworks {
		expectedAddresses := []string{}
		for _, port := range expectedStack {
			expectedAddresses = append(expectedAddresses, fmt.Sprintf(":%d", port))
		}

		actualStack := rtc.nr.stacks[netID].asList()

		assert.Len(rtc.t, actualStack, len(expectedAddresses))
		assert.Subset(rtc.t, actualStack, expectedAddresses)
	}
	return rtc
}

// Asserts that the pods have the existing connections. Does not guarantee that there are no other connections, i.e. the expected connections expected to be a subset of the actual connections, but not vice versa.
// This is because the mock peers used for these tests will not update their networks when peers fail. So we mainly test that the reconciler creates connections. It is not however resposible for deleteing the connections when peers are unavailable.
func (rtc *reconcilerTestCase) assertPodConnections(expectedNetworks map[int32][]int32) *reconcilerTestCase {
	for podPort, expectedNetwork := range expectedNetworks {
		expectedAddresses := []string{}
		for _, port := range expectedNetwork {
			expectedAddresses = append(expectedAddresses, fmt.Sprintf(":%d", port))
		}
		actualNetwork := rtc.testPods[podPort].lpNetwork
		log.Println(actualNetwork)

		assert.Subset(rtc.t, actualNetwork, expectedAddresses)
	}
	return rtc
}
