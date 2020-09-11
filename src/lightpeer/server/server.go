package server

import (
	"fmt"

	"google.golang.org/grpc"

	lpack "github.com/stefanprisca/lightchain/lightpeer"
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	grpctrace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc"
	"go.opentelemetry.io/otel/api/global"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func NewLPGrpcServer(host string, port int, blockRepo string) (*grpc.Server, *lpack.Lightpeer, lpack.NetworkHealthChecker) {
	peerAddress := fmt.Sprintf("%s:%d", host, port)
	tr := global.Tracer(fmt.Sprintf("%s-server@%s", lpack.ServiceName, peerAddress))
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(tr)),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(tr)))

	meta := pb.PeerInfo{Address: peerAddress}

	lp := &lpack.Lightpeer{
		Tracer:      tr,
		StoragePath: blockRepo,
		Meta:        meta,
		Network:     []pb.PeerInfo{meta},
	}

	pb.RegisterLightpeerServer(grpcServer, lp)
	healthpb.RegisterHealthServer(grpcServer, lp)

	nhc := lpack.NetworkHealthChecker{Lp: lp}
	nhc.StartPeerHealthCheck()
	return grpcServer, lp, nhc
}
