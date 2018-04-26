package main

import (
	"log"
	"net"

	//"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/BabelTowerGame/network/tob"
	"google.golang.org/grpc/reflection"
	"fmt"
)

const (
	port = "0.0.0.0:16882"
)

// server is used to implement tob.ToBServer.
type server struct {
	nodes map[string]pb.ToB_SubscribeServer
}

func newServer() *server {
	return &server{
		nodes: make(map[string]pb.ToB_SubscribeServer),
	}
}

// SayHello implements helloworld.GreeterServer
func (s *server) Subscribe(node *pb.NodeInfo, stream pb.ToB_SubscribeServer) error {
	id := node.GetId()
	fmt.Printf("Subscribe: node %v\n", id)
	s.nodes[id] = stream
	for {
		// long-lived stream
	}
	return nil
}

func (s *server) Publish(stream pb.ToB_PublishServer) error {
	for {
		event, err := stream.Recv()
		if err != nil {
			return err
		}
		switch event.GetTopic() {
		case pb.EventTopic_SERVER_EVENT:
		default:
			s.broadcast(event)
		}
	}
	return nil
}

func (s *server) broadcast(event *pb.Event) error {
	for _, stream := range s.nodes {
		stream.Send(event)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	server := newServer()
	pb.RegisterToBServer(s, server)

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
