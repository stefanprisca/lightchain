module github.com/stefanprisca/lightchain/src/lightpeer

go 1.15

replace (
	github.com/stefanprisca/lightchain => ../../
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v0.11.0
)

require (
	github.com/google/uuid v1.1.2
	github.com/stefanprisca/lightchain v0.0.0-20200930090534-72e6139961be
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc v0.11.0
	go.opentelemetry.io/otel v0.12.0
	go.opentelemetry.io/otel/exporters/otlp v0.11.0
	go.opentelemetry.io/otel/exporters/stdout v0.11.0
	go.opentelemetry.io/otel/sdk v0.11.0
	google.golang.org/grpc v1.32.0
)
