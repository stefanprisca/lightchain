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

// In self-recovery mode, a failed peer should be able to regain access to the network from
// the information stored in its persistant storage. e.g. recreate the chain from the blocks it knows about
// This is a requirement, since the applications using lightchain cannot have the full responsibility of
// maintaining the network. Client applications have the responsibility of joining peers together, but the peers
// themselves should do everything possible afterwards to maintain the network topology, which means rejoining
// the network in case of failure.

// The recovery module provides the recovery service a peer should use when starting.
// The steps for a successfull recovery are:
// 1) check the block storage for any blocks which were previously stored
// 	2) if no block stored, nothing to recover from so start a fresh peer
// 3) Iterate through the blocks and find the latest network block
// 4) Issue a join request on the peers from the known network
// 5) If none of the peers answer, the recovery failed => start new peer
// 6) If one peer answers, store the received chain and recovery was successful
