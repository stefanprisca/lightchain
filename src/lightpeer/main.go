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

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/stdout"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
)

func initOtel() {
	exporter, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
	if err != nil {
		log.Fatal(err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter),
	)

	if err != nil {
		log.Fatal(err)
	}

	global.SetTraceProvider(tp)
}

func main() {
	var verbose = flag.Bool("v", false, "runs verbose - gathering traces with otel")
	flag.Parse()

	if *verbose {
		initOtel()
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9081))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tr := global.Tracer("lightpeer")

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(tr)),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(tr)),
	)
	pb.RegisterLightpeerServer(grpcServer, &lightpeer{
		tr:          tr,
		storagePath: "testdata",
	})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
