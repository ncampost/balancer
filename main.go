package main

import (
	"fmt"
	"./system"
	"strconv"
	"time"
)

const NUM_WORKERS = 5

func main() {
	// Make Balancer machine
	balancer := "/tmp/balancer.sock"
	blr := system.MakeBalancer(balancer)

	// Make NUM_WORKERS worker machines, registering on Balancer
	workers := []*system.Worker{}
	for i := 0; i < NUM_WORKERS; i++ {
		workers = append(workers, system.MakeWorker("/tmp/worker" + strconv.Itoa(i) + ".sock", balancer))
	}

	// Sum all of the digits in nums.txt, distributing 30000 byte chunks to the worker machines.
	start := time.Now()
	sum := blr.DoAddJob("nums.txt", 30000)
	elapsed := time.Since(start)
	fmt.Printf("The sum of all the digits in nums.txt is %v.\n Time elapsed: %v\n", sum, elapsed)

	// Cleanup RPC servers, free Unix sockets
	blr.EndRPCServerUnix()
	for _, k := range workers {
		k.EndRPCServerUnix()
	}
}