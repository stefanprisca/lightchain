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
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/plugin/grpctrace"
)

func main() {
	var verbose = flag.Bool("v", false, "runs verbose - gathering traces with otel")
	var blockRepo = flag.String("repo", "testdata", "repo for storing the generated blocks")
	var otlpBackend = flag.String("otlp", OTLPAddress, "backend address for otlp traces and metrics")
	var host = flag.String("host", "", "the host to listen to")
	var port = flag.Int("port", 9081, "the port")
	flag.Parse()

	log.Printf("Starting the lightpeer with options: v: %v ; repo: %s ; otlp: %s\n",
		*verbose, *blockRepo, *otlpBackend)

	if *verbose {
		otelFinalizer := initOtel(*otlpBackend, ServiceName)
		defer otelFinalizer()
	}

	listenerAddress := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", listenerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tr := global.Tracer(ServiceName)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(tr)),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(tr)),
	)

	peerAddress := fmt.Sprintf("%s:%d", *host, *port)
	meta := pb.PeerInfo{Address: peerAddress}
	pb.RegisterLightpeerServer(grpcServer, &lightpeer{
		tr:          tr,
		storagePath: *blockRepo,
		meta:        meta,
		network:     []pb.PeerInfo{meta},
	})

	log.Println("Start serving gRPC connections...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
