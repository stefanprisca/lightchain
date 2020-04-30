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
	"io"
	"log"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/stdout"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"
)

func initOtel() {
	exporter, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
	if err != nil {
		log.Fatal(err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter),
	)

	if err != nil {
		log.Fatal(err)
	}

	global.SetTraceProvider(tp)
}

func main() {
	initOtel()

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9081", grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(global.Tracer(""))))

	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewLightpeerClient(conn)

	err = persistMessages(client)
	if err != nil {
		log.Fatalf("could not persist message: %v", err)
	}

	_, err = readMessages(client)
	if err != nil {
		log.Fatalf("could not read message: %v", err)
	}
}

func persistMessages(client pb.LightpeerClient) error {

	ctxt := context.Background()
	messages := []string{
		"Hello", "from", "the", "test", "side!",
	}

	for _, msg := range messages {
		persistReq := &pb.PersistRequest{
			Payload: []byte(msg),
		}
		_, err := client.Persist(ctxt, persistReq)
		if err != nil {
			return err
		}
	}
	return nil
}

func readMessages(client pb.LightpeerClient) (string, error) {

	ctxt := context.Background()

	queryClient, err := client.Query(ctxt, &pb.EmptyQueryRequest{})
	if err != nil {
		return "", err
	}

	log.Println("reading the message stream...")
	for {
		msg, err := queryClient.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.Query(_) = _, %v", client, err)
		}
		log.Println(string(msg.Payload))
	}

	return "", err
}
