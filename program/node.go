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
	chg       sync.Mutex
}

func (s *Node) init() {
	s.time = 0
	s.state = RELEASED
	s.queue = make(chan bool)
}

func (s *Node) Connect() {
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

// Nodes must agree to enter this critical section in turn.
var critical sync.Mutex

func (s *Node) Run() {
	for {
		s.enter()
		if critical.TryLock() {
			log.Printf("%v LOCK\n", s.Number)
			time.Sleep(60 * time.Millisecond)
			critical.Unlock()
			log.Printf("%v unlock\n", s.Number)
		} else {
			log.Panicf("%v Could not lock!\n", s.Number)
		}
		s.exit()
	}
}

func (s *Node) Request(ctx context.Context, msg *proto.Message) (*proto.Empty, error) {
	s.chg.Lock()
	x := s.state == HELD || s.state == WANTED && s.comesAfterMe(msg)
	s.time = max(s.time, msg.Time)
	s.chg.Unlock()

	if x {
		<-s.queue
	}
	return &proto.Empty{}, nil
}

func (s *Node) enter() {
	s.chg.Lock()
	s.state = WANTED
	msg := proto.Message{Time: s.time, Process: s.Number}
	s.chg.Unlock()

	// TODO: Jeg skal lægge til den logiske tid, så jeg er bag i køen ift. eventuelle imødekommende requests.
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
	s.time++
	s.chg.Unlock()
}

func (s *Node) exit() {
	s.chg.Lock()
	s.state = RELEASED
	s.chg.Unlock()
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
