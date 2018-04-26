package main

import (
	"google.golang.org/grpc"
	"log"

	pb "github.com/BabelTowerGame/network/tob"
	"context"
	"sync"
)

var (
	eventChan = make(chan *pb.Event, 10)
)

func subscribeWorker(client pb.ToBClient, wg *sync.WaitGroup) {
	defer wg.Done()

	stream, err := client.Subscribe(context.Background(), &pb.NodeInfo{
		Id: "node-id-fake",
	})
	if err != nil {
		log.Fatalf("Subscribe gRPC failed: %v\n", err)
	}

	for {
		event, err := stream.Recv()
		if err != nil {
			log.Printf("Subscribe stream Recv failed: %v\n", err)
		}

		log.Printf("Received event: %v\n", event)
	}
}

func publishWroker(client pb.ToBClient, wg *sync.WaitGroup) {
	defer wg.Done()

	stream, err := client.Publish(context.Background())
	if err != nil {
		log.Fatalf("Publish gRPC failed: %v\n", err)
	}

	for event := range eventChan {
		stream.Send(event)
	}
}

func main() {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	// Setup a connection with the server
	conn, err := grpc.Dial("127.0.0.1:16882", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}

	client := pb.NewToBClient(conn)

	go subscribeWorker(client, wg)
	go publishWroker(client, wg)
	wg.Add(2)
}
