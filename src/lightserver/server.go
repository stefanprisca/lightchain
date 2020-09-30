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

	"google.golang.org/grpc"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	lpack "github.com/stefanprisca/lightchain/src/lightpeer"
	grpctrace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc"
	"go.opentelemetry.io/otel/api/global"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func NewLPGrpcServer(host string, port int, blockRepo string) (*grpc.Server, *lpack.Lightpeer, lpack.NetworkHealthChecker) {
	peerAddress := fmt.Sprintf("%s:%d", host, port)
	tr := global.Tracer(fmt.Sprintf("%s-server@%s", lpack.ServiceName, peerAddress))
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(tr)),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(tr)))

	meta := pb.PeerInfo{Address: peerAddress}

	lp := &lpack.Lightpeer{
		Tracer:      tr,
		StoragePath: blockRepo,
		Meta:        meta,
		Network:     []pb.PeerInfo{meta},
	}

	pb.RegisterLightpeerServer(grpcServer, lp)
	healthpb.RegisterHealthServer(grpcServer, lp)

	nhc := lpack.NetworkHealthChecker{Lp: lp}
	nhc.StartPeerHealthCheck()
	return grpcServer, lp, nhc
}
