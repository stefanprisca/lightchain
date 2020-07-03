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
	"log"

	"github.com/open-telemetry/opentelemetry-collector/translator/conventions"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

const (
	ServiceName = "lightchain"
	OTLPAddress = "localhost:30080"
)

func initOtel(otlpBackend, serviceName string) func() error {

	otlpExp, err := otlp.NewExporter(otlp.WithInsecure(),
		otlp.WithAddress(otlpBackend),
		otlp.WithGRPCDialOption(grpc.WithBlock()))
	if err != nil {
		log.Fatalf("Failed to create the collector exporter: %v", err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResourceAttributes(
			// the service name used to display traces in Jaeger
			kv.Key(conventions.AttributeServiceName).String(serviceName),
		),
		sdktrace.WithSyncer(otlpExp),
	)
	if err != nil {
		log.Fatalf("error creating trace provider: %v\n", err)
	}

	global.SetTraceProvider(tp)
	return func() error {
		global.SetTraceProvider(&apitrace.NoopProvider{})
		return otlpExp.Stop()
	}
}
