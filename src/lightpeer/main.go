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
	"google.golang.org/grpc"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/trace/stdout"
	"go.opentelemetry.io/otel/plugin/grpctrace"

	"github.com/open-telemetry/opentelemetry-collector/translator/conventions"
)

const (
	ServiceName = "lightchain"
	OTLPAddress = "localhost:30080"
)

func initOtel() func() error {
	stdOutExp, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
	if err != nil {
		log.Fatal(err)
	}

	otlpExp, err := otlp.NewExporter(otlp.WithInsecure(),
		otlp.WithAddress(OTLPAddress),
		otlp.WithGRPCDialOption(grpc.WithBlock()))
	if err != nil {
		log.Fatalf("Failed to create the collector exporter: %v", err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(stdOutExp),
		sdktrace.WithResourceAttributes(
			// the service name used to display traces in Jaeger
			kv.Key(conventions.AttributeServiceName).String(ServiceName),
		),
		sdktrace.WithBatcher(otlpExp, // add following two options to ensure flush
			sdktrace.WithScheduleDelayMillis(5),
			sdktrace.WithMaxExportBatchSize(2),
		),
	)
	if err != nil {
		log.Fatalf("error creating trace provider: %v\n", err)
	}

	global.SetTraceProvider(tp)
	return func() error {
		return otlpExp.Stop()
	}
}

func main() {
	var verbose = flag.Bool("v", false, "runs verbose - gathering traces with otel")
	var blockRepo = flag.String("repo", "testdata", "repo for storing the generated blocks")
	flag.Parse()

	if *verbose {
		otelFinalizer := initOtel()
		defer otelFinalizer()
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9081))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tr := global.Tracer(ServiceName)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(tr)),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(tr)),
	)
	pb.RegisterLightpeerServer(grpcServer, &lightpeer{
		tr:          tr,
		storagePath: *blockRepo,
	})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
