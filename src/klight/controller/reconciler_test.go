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
	"testing"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReconcilerStacksIPs(t *testing.T) {
	rtc := newReconcilerTestCase(t)
	rtc.addPod("10.0.0.1", "networkId").
		addPod("10.0.0.2", "networkId").
		addPod("10.0.0.3", "networkId4").
		addPod("10.0.0.4", "networkId2").
		assertStacks(map[string][]string{
			"networkId":  {"10.0.0.2", "10.0.0.1"},
			"networkId4": {"10.0.0.3"},
			"networkId2": {"10.0.0.4"},
		})
}

type reconcilerTestCase struct {
	nr *networkReconciler
	t  *testing.T
}

func newReconcilerTestCase(t *testing.T) reconcilerTestCase {
	return reconcilerTestCase{
		nr: &networkReconciler{map[string]ipStack{}},
		t:  t,
	}
}

func (rtc reconcilerTestCase) addPod(ip, networkId string) reconcilerTestCase {
	pod := &v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{Labels: map[string]string{klightNetworkLabel: networkId}},
		Status:     v1.PodStatus{PodIP: ip},
	}

	rtc.nr.reconcileLightNetwork(pod)
	return rtc
}

func (rtc reconcilerTestCase) assertStacks(expectedStacks map[string][]string) {
	for netId, expectedStack := range expectedStacks {
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
