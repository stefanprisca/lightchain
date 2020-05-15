module github.com/stefanprisca/lightchain

go 1.14

// replace go.opentelemetry.io/otel => github.com/stefanprisca/opentelemetry-go v0.4.4-0.20200430143930-c3e9bdb214a6

require (
	github.com/golang/protobuf v1.4.0-rc.2
	github.com/google/uuid v1.1.1
	github.com/open-telemetry/opentelemetry-collector v0.3.0
	go.opentelemetry.io/otel v0.5.0
	go.opentelemetry.io/otel/exporters/otlp v0.5.0
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.20.0 // indirect
)
