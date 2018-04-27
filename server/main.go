package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	pb "github.com/BabelTowerGame/network/tob"
	"google.golang.org/grpc/reflection"
	"fmt"
	"io"
	"sync"
	"google.golang.org/grpc/metadata"
	"errors"
	"math/rand"
)

const (
	port = "0.0.0.0:16882"
)

// server is used to implement tob.ToBServer.
type server struct {
	nodes           map[string]pb.ToB_SubscribeServer
	serverNode      string
	serverNodeMutex sync.Mutex
}

func newServer() *server {
	return &server{
		nodes:      make(map[string]pb.ToB_SubscribeServer),
		serverNode: "",
	}
}

// SayHello implements helloworld.GreeterServer
func (s *server) Subscribe(_ *pb.Empty, stream pb.ToB_SubscribeServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("fail to get metadata")
	}
	ids := md.Get("id")
	if len(ids) < 1 || ids[0] == "" {
		return errors.New("empty node ID")
	}
	id := ids[0]


	fmt.Printf("Subscribe: node %v\n", id)

	// Register the node
	s.nodes[id] = stream

	stream.Context()

	// Update server node if needed
	s.serverNodeMutex.Lock()
	if s.serverNode == "" {
		// server node not assigned
		// assign server node to this new node
		s.serverNode = id
		// Tell all nodes about this event
		s.broadcast(&pb.Event{
			Topic: pb.EventTopic_SERVER_EVENT,
			S: &pb.ServerEvent{
				Id: id,
				Type: pb.ServerEventType_SERVER_CHANGE,
			},
		}, false)
	}
	s.serverNodeMutex.Unlock()

	// Tell the node who's the current server
	// So that it knows whether to use server logic or not
	stream.Send(&pb.Event{
		Topic: pb.EventTopic_SERVER_EVENT,
		S: &pb.ServerEvent{
			Id: s.serverNode,
			Type: pb.ServerEventType_SERVER_CHANGE,
		},
	})

	// Long-live stream
	// Keep alive until the client disconnects
	<-stream.Context().Done()

	// Client disconnected, so we un-register the node
	delete(s.nodes, id)
	fmt.Printf("Un-Subscribe: node %v\n", id)

	if id == s.serverNode {
		// The server is down, so we need to assign a new one
		if len(s.nodes) > 0 {
			// Randomly choose a registered node
			i := rand.Intn(len(s.nodes))
			var newServer string
			for newServer = range s.nodes {
				if i == 0 {
					break
				}
				i--
			}
			// Set the node as the new server node
			s.serverNode = newServer
			// Tell all nodes the new server
			s.broadcast(&pb.Event{
				Topic: pb.EventTopic_SERVER_EVENT,
				S: &pb.ServerEvent{
					Id: newServer,
					Type: pb.ServerEventType_SERVER_CHANGE,
				},
			}, true)
		}
	}
	return nil
}

func (s *server) Publish(stream pb.ToB_PublishServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("fail to get metadata")
	}
	ids := md.Get("id")
	if len(ids) < 1 || ids[0] == "" {
		return errors.New("empty node ID")
	}
	id := ids[0]

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			// Oh! the client ended the stream
			// It's very likely the client shuts down
		}
		if err != nil {
			return err
		}
		fmt.Printf("Received from %v: %v\n", id, event)
		switch event.GetTopic() {
		case pb.EventTopic_SERVER_EVENT:
			// No one should publish server event
			// Ignore the event
		default:
			if id == s.serverNode {
				s.broadcast(event, false)
				fmt.Printf("Broadcast to all nodes\n")
			} else {
				s.nodes[s.serverNode].Send(event)
				fmt.Printf("Forward to server %v\n", s.serverNode)
			}
		}
	}
	return nil
}

func (s *server) broadcast(event *pb.Event, includeServer bool) error {
	for id, stream := range s.nodes {
		if includeServer || id != s.serverNode {
			stream.Send(event)
		}
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
