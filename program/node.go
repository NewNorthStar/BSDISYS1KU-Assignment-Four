package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	proto "example.com/ricard/grpc"
	"google.golang.org/grpc"
)

var myProcessNumber int64 = time.Now().Unix()
var myTime int64 = 1
var myState state = RELEASED
var queue chan bool = make(chan bool)

type state int64

const (
	RELEASED state = 1
	WANTED   state = 2
	HELD     state = 3
)

type Node struct {
	proto.UnimplementedRicardServiceServer
	Address   string
	Instances []string
}

func (s *Node) Launch() {
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
	if myState == HELD || myState == WANTED && comesAfterMe(msg) {
		<-queue
	}
	return &proto.Empty{}, nil
}

func enter() {
	myState = WANTED
	var replies sync.WaitGroup

	for i := 0; i < 3; i++ {
		replies.Add(1)

		go func() {
			defer replies.Done()
			// Contact other node
		}()
	}

	replies.Wait()
	myState = HELD
}

func exit() {
	myState = RELEASED
main:
	for {
		select {
		case queue <- true:
		default:
			break main
		}
	}
}

func comesAfterMe(msg *proto.Message) bool {
	if myTime != msg.Time {
		return myTime < msg.Time
	} else {
		return myProcessNumber < msg.Process
	}
}
