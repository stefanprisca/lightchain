module github.com/stefanprisca/lightchain/klight/klightpeer

go 1.15

// replace github.com/stefanprisca/lightchain/src/lightpeer => ../../lightpeer

require (
	github.com/stefanprisca/lightchain v0.0.0-20200929093804-5c21f182115a
	github.com/stefanprisca/lightchain/src/lightpeer v0.0.0-20200929093804-5c21f182115a
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc v0.11.0
	go.opentelemetry.io/otel v0.11.0
	google.golang.org/grpc v1.32.0
)
