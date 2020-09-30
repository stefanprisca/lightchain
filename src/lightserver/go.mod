module github.com/stefanprisca/lightchain/src/lightserver

go 1.15

replace (
	github.com/stefanprisca/lightchain => ../../
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v0.11.0
)

require (
	github.com/google/uuid v1.1.2
	github.com/stefanprisca/lightchain v0.0.0-20200930090534-72e6139961be
	github.com/stefanprisca/lightchain/src/lightpeer v0.0.0-20200929093804-5c21f182115a
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc v0.11.0
	go.opentelemetry.io/otel v0.12.0
	go.opentelemetry.io/otel/exporters/stdout v0.11.0 // indirect
	google.golang.org/grpc v1.32.0
)
