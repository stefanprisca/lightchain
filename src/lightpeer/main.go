package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9081))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterLightpeerServer(grpcServer, &lightpeer{})

	grpcServer.Serve(lis)
}
