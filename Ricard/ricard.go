package main

import (
	"context"
	"sync"
	"time"

	proto "example.com/ricard/grpc"
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

type RicardServer struct {
	proto.UnimplementedRicardServiceServer
}

func main() {

}

func (s *RicardServer) Request(ctx context.Context, msg *proto.Message) (*proto.Empty, error) {
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
