package main

import (
	"context"
	"io"
	"log"

	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
	"google.golang.org/grpc"
)

func main() {
	serverAddr := "localhost:9081"
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to server at %s, %v", serverAddr, err)
	}
	defer conn.Close()

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
	messages := []string{
		"Hello", "from", "the", "test", "side!",
	}
	ctxt := context.Background()
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
