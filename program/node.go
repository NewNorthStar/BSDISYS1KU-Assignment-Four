package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	proto "example.com/ricard/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
Enumerate representing the logical state of access to the CS.

As shown in the Ricart Agrawala Algorithm example
given at lecture 7, Distributed Systems E2024 at ITU.
*/
type state int64

const (
	RELEASED state = 1
	WANTED   state = 2
	HELD     state = 3
)

/*
Each node runs as a coroutine, with the resources and states in this struct.
*/
type Node struct {
	proto.UnimplementedRicardServiceServer
	Number      int64
	Address     string
	Instances   []string
	time        int64
	state       state
	replySignal chan bool
	replyGroup  sync.WaitGroup
	chg         sync.Mutex
	log         log.Logger
}

/*
Setup default values for a node.
*/
func (s *Node) init() {
	s.time = 0
	s.state = RELEASED
	s.replySignal = make(chan bool)

	file, err := os.Create(strconv.FormatInt(s.Number, 10) + ".txt")
	if err != nil {
		log.Fatalf(err.Error())
	}
	s.log.SetOutput(file)
}

//// SERVER SIDE FUNCTIONS ////

/*
Setup and run the server side of the node as a coroutine.
*/
func (s *Node) Connect() {
	s.init()
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		s.log.Fatalf("Failed to establish listener: %v\n", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterRicardServiceServer(grpcServer, s)
	s.log.Printf("%v ready for service.\n", listener.Addr())

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			s.log.Fatalf("failed to establish service: %v\n", err)
		}
	}()
}

/*
Handles replies to other nodes.
*/
func (s *Node) Request(ctx context.Context, msg *proto.Message) (*proto.Empty, error) {
	defer s.log.Printf("%v replied to %v", s.Number, msg.Process)
	s.chg.Lock()
	x := s.state == HELD || s.state == WANTED && s.comesAfterMe(msg)
	s.chg.Unlock()

	if x {
		s.replyGroup.Add(1)
		defer s.replyGroup.Done()
		<-s.replySignal
	}

	s.chg.Lock()
	s.time = max(s.time, msg.Time)
	s.chg.Unlock()
	return &proto.Empty{}, nil
}

/*
Compares a message to this node.
In the case where both want to access the CS.
*/
func (s *Node) comesAfterMe(msg *proto.Message) bool {
	if s.time != msg.Time {
		return s.time < msg.Time
	} else {
		return s.Number < msg.Process
	}
}

//// CLIENT SIDE FUNCTIONS ////

/*
Runs the client side of the node as a coroutine.
*/
func (s *Node) Run() {
	go s.clientSideRoutine()
}

/*
Nodes must enter this critical section legally.
*/
var critical sync.Mutex

/*
Work sequence of the node client side.

The node will continuously attempt to reach the CS, and panics if it tries an illegal access.
*/
func (s *Node) clientSideRoutine() {
	for {
		s.enter()
		if critical.TryLock() {
			s.log.Printf("%v LOCK\n", s.Number)
			time.Sleep(1000 * time.Millisecond)
			critical.Unlock()
			s.log.Printf("%v unlock\n", s.Number)
		} else {
			s.log.Panicf("%v Could not lock!\n", s.Number)
		}
		s.exit()
	}
}

/*
Secure permission to enter the critical section.
*/
func (s *Node) enter() {
	s.chg.Lock()
	s.state = WANTED
	msg := &proto.Message{Time: s.time, Process: s.Number}
	s.chg.Unlock()

	var replies sync.WaitGroup

	for i := 0; i < len(s.Instances); i++ {
		if s.Instances[i] == s.Address { // No self-requests
			continue
		}

		replies.Add(1)
		go func() {
			defer replies.Done()
			s.obtainLockFromPeer(i, msg)
		}()
	}

	replies.Wait()

	s.chg.Lock()
	s.state = HELD
	s.chg.Unlock()
}

/*
Request permission to access the critical section from peer.
*/
func (s *Node) obtainLockFromPeer(index int, msg *proto.Message) {
	conn, err := grpc.NewClient(s.Instances[index], grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			s.log.Fatalf("%v Failed to close grpcClientConn.", s.Number)
		}
	}(conn)
	if err != nil {
		s.log.Fatalf("Failed to obtain connection: %v\n", err)
	}

	s.log.Printf("%v requesting %v, time %v\n", s.Number, index, s.time)

	_, err = proto.NewRicardServiceClient(conn).Request(context.Background(), msg)
	if err != nil {
		s.log.Fatalf("Failed to obtain connection: %v\n", err)
	}

	s.log.Printf("%v got return from %v\n", s.Number, index)
}

/*
Release lock on the critical section. Reply to waiting nodes.

Waits for replies to complete. Ensures that this node
will then have a Lamport time later than any nodes it replied to.
*/
func (s *Node) exit() {
main:
	for {
		select {
		case s.replySignal <- true:
		default:
			break main
		}
	}
	s.replyGroup.Wait()
	s.chg.Lock()
	s.state = RELEASED
	s.time++
	s.chg.Unlock()
}
