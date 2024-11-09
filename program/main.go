package main

import (
	"strconv"
	"time"
)

const N = 5

func main() {
	instances := make([]string, N, N)
	for i := 0; i < N; i++ {
		instances[i] = "localhost:" + strconv.Itoa(5050+i)
	}
	nodes := make([]Node, N, N)
	for i := int64(0); i < N; i++ {
		nodes[i] = Node{
			Number:    i,
			Address:   instances[i],
			Instances: instances,
		}
		go nodes[i].Connect()
	}

	time.Sleep(time.Second)
}
