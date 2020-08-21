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
	"fmt"
	"log"

	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
)

type ipStack struct {
	ips []string
}

func (is ipStack) pushIp(ip string) ipStack {
	is.ips = append([]string{ip}, is.ips...)
	return is
}

func (is ipStack) popIp() (string, ipStack, error) {
	if len(is.ips) == 0 {
		return "", is, fmt.Errorf("stack is empty")
	}

	ip := is.ips[0]
	is.ips = is.ips[1:]
	return ip, is, nil
}

func (is ipStack) lookUp() (string, error) {
	if len(is.ips) == 0 {
		return "", fmt.Errorf("stack is empty")
	}

	return is.ips[0], nil
}

func (is ipStack) asList() []string {
	return is.ips
}

type networkReconciler struct {
	stacks map[string]ipStack
}

// Reconciling should be as stateless as possible, as k8s pods are volatile and there are no guarantees of what's up and what's down. But at the same time it needs to keep track of contact pods for each network id, such that new pods can join the network if it already exists.
// This can be done by maintaining an IP stack for each network id, with the newest known pod at the top of the stack. When a new pod (podA) wants to join the network, reconciliation works as follows:
// 1) stack is empty => podA is the only known one in the network, so push podA.IP to stack
// 2) stack is not empty => read, without poping, the first IP on the stack and try to connect podA to that IP.
// 2.1) Connection is successful => push podA.IP to the stack
// 2.2) Connection unsuccessful => pop the head of the stack, cleaning up down pods, and jump to step 1.

// This method should ensure that all known pods for a network are recorded, and if there is one alive on that network, then new pods will be able to join it. And since it is poping the existing nodes in case of unsuccessful connections, the stack should be pretty clean (although there can still be leftovers at the bottom which need cleaning).

func (nr *networkReconciler) reconcileLightNetwork(pod *v1.Pod) {

	log.Println(pod.Status.Phase)

	podIp := pod.Status.PodIP
	log.Println(podIp)

	networkId, ok := pod.Labels[klightNetworkLabel]
	if !ok {
		klog.Fatal("pod does not have klight network id")
	}

	log.Println(networkId)

	newStack := nr.stacks[networkId].pushIp(podIp)
	nr.stacks[networkId] = newStack
}
