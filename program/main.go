package main

import (
	"fmt"
	"strconv"
	"time"
)

const n = 20

func main() {
	addresses := createAddressList()
	nodes := createNodes(addresses)
	time.Sleep(time.Second) // Grace period for node service listener startup.
	runAllNodes(nodes)
	time.Sleep(time.Second * 30) // Duration of run.
}

/*
Creates a list of localhost addresses which nodes will be connected to.

Supplied to all the nodes as a stand in for service discovery.
*/
func createAddressList() []string {
	list := make([]string, 0, n)
	for i := 0; i < n; i++ {
		a := "localhost:" + strconv.Itoa(5050+i)
		list = append(list, a)
		fmt.Println(a)
	}
	return list
}

/*
Creates the nodes, connected to their addresses and with the gRPC service running.
*/
func createNodes(addresses []string) []*Node {
	nodes := make([]*Node, 0, n)
	for i := int64(0); i < n; i++ {
		node := Node{
			Number:    i,
			Address:   addresses[i],
			Instances: addresses,
		}
		nodes = append(nodes, &node)
		node.Connect()
	}
	return nodes
}

/*
Starts the client side 'Run()' function of all nodes as a coroutine.
*/
func runAllNodes(nodes []*Node) {
	for _, node := range nodes {
		node.Run()
	}
}
