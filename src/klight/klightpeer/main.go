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
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	lpack "github.com/stefanprisca/lightchain/src/lightpeer"
	grpctrace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc"
	"go.opentelemetry.io/otel/api/global"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	var verbose = flag.Bool("v", false, "runs verbose - gathering traces with otel")
	var blockRepo = flag.String("repo", "testdata", "repo for storing the generated blocks")
	var otlpBackend = flag.String("otlp", lpack.OTLPAddress, "backend address for otlp traces and metrics")
	var host = flag.String("host", "", "the host to listen to")
	var port = flag.Int("port", 9081, "the port")

	var _ = flag.String("stateFile", "", "the path to the state file")
	flag.Parse()

	log.Printf("Starting the lightpeer with options: v: %v ; repo: %s ; otlp: %s\n",
		*verbose, *blockRepo, *otlpBackend)

	if *verbose {
		otelFinalizer := lpack.InitOtel(*otlpBackend, lpack.ServiceName)
		defer otelFinalizer()
	}
	listenerAddress := fmt.Sprintf("%s:%d", *host, *port)
	lis, err := net.Listen("tcp", listenerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	localIp, err := getLocalIP()
	if err != nil {
		log.Fatalf("failed to get ip: %v", err)
	}

	peerAddress := fmt.Sprintf("%s:%d", localIp, port)
	tr := global.Tracer(fmt.Sprintf("%s-server@%s", lpack.ServiceName, peerAddress))
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(tr)),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(tr)))

	meta := pb.PeerInfo{Address: peerAddress}
	lp := &lpack.Lightpeer{
		Tracer:      tr,
		StoragePath: *blockRepo,
		Meta:        meta,
		Network:     []pb.PeerInfo{meta},
	}

	klp := &klightpeer{lp}

	pb.RegisterLightpeerServer(grpcServer, klp)
	healthpb.RegisterHealthServer(grpcServer, klp)

	nhc := lpack.NetworkHealthChecker{Lp: lp}
	nhc.StartPeerHealthCheck()

	defer nhc.StopPeerHealthCheck()
	log.Println("Start serving gRPC connections @ ", listenerAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}

	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr).IP
	return fmt.Sprintf("%v", localAddr), nil
}
