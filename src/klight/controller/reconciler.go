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

	"google.golang.org/grpc"
	"k8s.io/klog"

	lpb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	v1 "k8s.io/api/core/v1"
)

type addressStack struct {
	addresses []string
}

func (addrStack addressStack) push(addr string) addressStack {
	addrStack.addresses = append([]string{addr}, addrStack.addresses...)
	return addrStack
}

func (addrStack addressStack) pop() (string, addressStack, error) {
	if len(addrStack.addresses) == 0 {
		return "", addrStack, fmt.Errorf("stack is empty")
	}

	addr := addrStack.addresses[0]
	addrStack.addresses = addrStack.addresses[1:]
	return addr, addrStack, nil
}

func (addrStack addressStack) lookUp() (string, bool) {
	if len(addrStack.addresses) == 0 {
		return "", false
	}

	return addrStack.addresses[0], true
}

func (addrStack addressStack) asList() []string {
	return addrStack.addresses
}

type networkReconciler struct {
	stacks map[string]addressStack
}

// Reconciling should be as stateless as possible, as k8s pods are volatile and there are no guarantees of what's up and what's down. But at the same time it needs to keep track of contact pods for each network id, such that new pods can join the network if it already exists.
// This can be done by maintaining an IP stack for each network id, with the newest known pod at the top of the stack. When a new pod (podA) wants to join the network, reconciliation works as follows:
// 1) stack is empty => podA is the only known one in the network, so push podA.IP to stack
// 2) stack is not empty => read, without poping, the first IP on the stack and try to connect podA to that IP.
// 2.1) Connection is successful => push podA.IP to the stack
// 2.2) Connection unsuccessful => pop the head of the stack, cleaning up down pods, and jump to step 1.

// This method should ensure that all known pods for a network are recorded, and if there is one alive on that network, then new pods will be able to join it. And since it is poping the existing nodes in case of unsuccessful connections, the stack should be pretty clean (although there can still be leftovers at the bottom which need cleaning).

func (nr *networkReconciler) reconcileLightNetwork(pod v1.Pod) error {

	networkId, ok := pod.Labels[klightNetworkLabel]
	if !ok {
		klog.Fatal("pod does not have klight network id")
	}

	podAddress := getPodAddress(pod)
	addressStack := nr.stacks[networkId]
	for {
		log.Printf("reconciling net for pod %s, with stack %v", podAddress, addressStack)
		existingAddress, ok := addressStack.lookUp()
		if !ok {
			break
		}

		if existingAddress == podAddress {
			_, addressStack, _ = addressStack.pop()
			continue
		}

		err := joinPodToNetwork(podAddress, existingAddress)
		if err == nil {
			break
		}
		log.Println(err)
		_, addressStack, _ = addressStack.pop()
	}

	addressStack = addressStack.push(podAddress)
	nr.stacks[networkId] = addressStack

	return nil
}

func getPodAddress(pod v1.Pod) string {
	podIp := pod.Status.PodIP
	var podLPPort int32 = 9081
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name != klightPodPort {
				continue
			}
			podLPPort = port.ContainerPort
		}
	}

	return fmt.Sprintf("%s:%d", podIp, podLPPort)
}

func joinPodToNetwork(podAddress, networkContactAddress string) error {
	log.Println("joining pods to network", podAddress, networkContactAddress)

	conn, err := grpc.Dial(podAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("did not connect: %s", err)
	}
	defer func() { _ = conn.Close() }()

	client := lpb.NewLightpeerClient(conn)
	joinReq := &lpb.JoinRequest{
		Address: networkContactAddress,
	}

	_, err = client.JoinNetwork(context.Background(), joinReq)

	log.Println("sent the request to join network")
	return err
}
