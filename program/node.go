package main

import (
	"context"
	"log"
	"net"
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
}

func (s *Node) init() {
	s.time = 0
	s.state = RELEASED
	s.replySignal = make(chan bool)
}

//// SERVER SIDE FUNCTIONS ////

/*
Sets up and runs the server side of the node as a coroutine.
*/
func (s *Node) Connect() {
	s.init()
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		log.Fatalf("Failed to establish listener: %v\n", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterRicardServiceServer(grpcServer, s)
	log.Printf("%v ready for service.\n", listener.Addr())

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			log.Fatalf("failed to establish service: %v\n", err)
		}
	}()
}

func (s *Node) Request(ctx context.Context, msg *proto.Message) (*proto.Empty, error) {
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
			log.Printf("%v LOCK\n", s.Number)
			time.Sleep(200 * time.Millisecond)
			critical.Unlock()
			log.Printf("%v unlock\n", s.Number)
		} else {
			log.Panicf("%v Could not lock!\n", s.Number)
		}
		s.exit()
	}
}

func (s *Node) enter() {
	s.chg.Lock()
	s.state = WANTED
	msg := proto.Message{Time: s.time, Process: s.Number}
	s.chg.Unlock()

	var replies sync.WaitGroup

	for i := 0; i < len(s.Instances); i++ {
		if s.Instances[i] == s.Address {
			continue
		}

		replies.Add(1)
		go func(address *string) {
			defer replies.Done()
			conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			defer func(conn *grpc.ClientConn) {
				err := conn.Close()
				if err != nil {
					log.Fatalf("%v Failed to close grpcClientConn.", s.Number)
				}
			}(conn)
			if err != nil {
				log.Fatalf("Failed to obtain connection: %v\n", err)
			}
			//log.Printf("%v requesting %s, time %v\n", s.Number, *address, s.time)
			_, err = proto.NewRicardServiceClient(conn).Request(context.Background(), &msg)
			if err != nil {
				log.Fatalf("Failed to obtain connection: %v\n", err)
			}
			//log.Printf("%v got return from %s\n", s.Number, *address)
		}(&s.Instances[i])
	}

	replies.Wait()
	s.chg.Lock()
	s.state = HELD
	s.chg.Unlock()
}

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
