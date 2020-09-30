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
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"
)

func main() {
	var lpAddress = flag.String("lpAddress", ":9081", "the address for the peer to connect to")
	flag.Parse()

	log.Println("Starting persister with peer address %s", *lpAddress)

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(*lpAddress, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewLightpeerClient(conn)

	err = persistMessages(client)
	if err != nil {
		log.Fatalf("could not persist message: %v", err)
	}
}

func persistMessages(client pb.LightpeerClient) error {

	ctxt := context.Background()

	for {
		time.Sleep(time.Minute)
		message := fmt.Sprintf("Hello from the persistor! it's %v o'clock!", time.Now())
		persistReq := &pb.PersistRequest{
			Payload: []byte(message),
		}
		_, err := client.Persist(ctxt, persistReq)
		if err != nil {
			return err
		}
		log.Println("sent message: %v", persistReq)
	}
	return nil
}
