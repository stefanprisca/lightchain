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
