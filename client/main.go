package main

import (
	"google.golang.org/grpc"
	"log"

	pb "github.com/BabelTowerGame/network/tob"
	"context"
)

var (
	eventChan = make(chan *pb.Event, 10)
)

func main() {
	// Setup a connection with the server
	conn, err := grpc.Dial("127.0.0.1:6882", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := pb.NewToBClient(conn)

	stream, err := client.Publish(context.Background())

	for event := range eventChan {
		stream.Send(event)
	}
}
