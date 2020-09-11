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

package lightpeer

import (
	"context"
	"fmt"
	"time"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	grpctrace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc"
	"go.opentelemetry.io/otel/api/global"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	nhcRunning = 1
	nhcStopped = 2
)

type NetworkHealthChecker struct {
	Lp     *Lightpeer
	status int
}

func (nhc *NetworkHealthChecker) StartPeerHealthCheck() {
	nhc.status = nhcRunning
	lp := nhc.Lp
	go func() {
		for nhc.status == nhcRunning {
			ctx := context.Background()
			nhcCtx, span := lp.Tracer.Start(ctx, fmt.Sprintf("@%s - Network healthcheck", lp.Meta.Address))

			oldNetwork := lp.Network
			newNetwork := []pb.PeerInfo{}

			for _, peer := range oldNetwork {

				if peer.Address == lp.Meta.Address {
					newNetwork = append(newNetwork, peer)
				} else if isAlive(nhcCtx, peer) {
					newNetwork = append(newNetwork, peer)
				}
			}

			if len(newNetwork) != len(oldNetwork) {
				lp.Network = newNetwork
				err := lp.updateNetwork(nhcCtx, newNetwork)
				if err != nil {
					lp.Network = oldNetwork
				}
			}
			span.End()

			time.Sleep(500 * time.Millisecond)
		}

	}()
}

func isAlive(nhcCtx context.Context, peer pb.PeerInfo) bool {
	conn, err := grpc.Dial(peer.Address, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(
			global.Tracer(fmt.Sprintf("client@%s", peer.Address)))),
		grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(
			global.Tracer(fmt.Sprintf("stream-client@%s", peer.Address)))))
	defer conn.Close()
	if err != nil {
		return false
	}

	client := healthpb.NewHealthClient(conn)
	resp, err := client.Check(nhcCtx, &healthpb.HealthCheckRequest{})
	if err != nil || resp.Status != healthpb.HealthCheckResponse_SERVING {
		return false
	}
	return true
}

func (nhc *NetworkHealthChecker) StopPeerHealthCheck() {
	nhc.status = nhcStopped
}
