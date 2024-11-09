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
	for i := 0; i < N; i++ {
		nodes[i] = Node{
			Address:   instances[i],
			Instances: instances,
		}
		go nodes[i].Launch()
	}

	time.Sleep(time.Second)
}
