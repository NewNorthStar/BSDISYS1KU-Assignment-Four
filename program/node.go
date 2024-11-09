package main

import (
	"context"
	"log"
	"net"
	"sync"

	proto "example.com/ricard/grpc"
	"google.golang.org/grpc"
)

type state int64

const (
	RELEASED state = 1
	WANTED   state = 2
	HELD     state = 3
)

type Node struct {
	proto.UnimplementedRicardServiceServer
	Number    int64
	Address   string
	Instances []string
	time      int64
	state     state
	queue     chan bool
}

func (s *Node) init() {
	s.time = 1
	s.state = RELEASED
	s.queue = make(chan bool)
}

func (s *Node) Launch() {
	s.init()
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		log.Fatalf("Failed to establish listener: %v\n", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterRicardServiceServer(grpcServer, s)
	log.Printf("%v ready for service.\n", listener.Addr())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("failed to establish service: %v\n", err)
	}
}

func (s *Node) RunTasks() {

}

func (s *Node) Request(ctx context.Context, msg *proto.Message) (*proto.Empty, error) {
	if s.state == HELD || s.state == WANTED && s.comesAfterMe(msg) {
		<-s.queue
	}
	return &proto.Empty{}, nil
}

func (s *Node) enter() {
	s.state = WANTED
	var replies sync.WaitGroup

	for i := 0; i < 3; i++ {
		replies.Add(1)

		go func() {
			defer replies.Done()
			// Contact other node
		}()
	}

	replies.Wait()
	s.state = HELD
}

func (s *Node) exit() {
	s.state = RELEASED
main:
	for {
		select {
		case s.queue <- true:
		default:
			break main
		}
	}
}

func (s *Node) comesAfterMe(msg *proto.Message) bool {
	if s.time != msg.Time {
		return s.time < msg.Time
	} else {
		return s.Number < msg.Process
	}
}
