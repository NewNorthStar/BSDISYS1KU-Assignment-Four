package main

import (
	"context"
	"time"

	proto "example.com/ricard/grpc"
)

var myProcessNumber int64 = time.Now().Unix()
var myTime int64 = 1
var myState state = RELEASED

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

}
