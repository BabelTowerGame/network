package main

import (
	"google.golang.org/grpc"
	"log"

	pb "github.com/BabelTowerGame/network/tob"
	"context"
	"sync"
	"math/rand"
	"time"
	"google.golang.org/grpc/metadata"
)

var (
	id string
	eventChan = make(chan *pb.Event, 10)
)

func init() {
	rand.Seed(time.Now().UnixNano())
	id = RandStringRunes(32)
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func subscribeWorker(client pb.ToBClient, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "id", id)

	stream, err := client.Subscribe(ctx, &pb.Empty{})
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

	ctx := metadata.AppendToOutgoingContext(context.Background(), "id", id)

	stream, err := client.Publish(ctx)
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
