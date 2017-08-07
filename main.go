package main

import (
	"fmt"
	"./system"
	"strconv"
	"time"
	"os"
)

const NUM_WORKERS = 10

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
	/*
	start := time.Now()
	sum := blr.DoAddJob("nums.txt", 30000)
	elapsed := time.Since(start)
	fmt.Printf("The sum of all the digits in nums.txt is %v.\nTime elapsed: %v\n", sum, elapsed)
	*/

	file, _ := os.OpenFile("results.txt", os.O_WRONLY | os.O_APPEND, 0666)
	var elapsedavg float64 = 0
	for i := 0; i < 10; i++ {
		fmt.Printf("Starting trial %v.\n", i+1)
		start := time.Now()
		sum := blr.DoAddJob("nums.txt", 30000)
		elapsed := time.Since(start)
		fmt.Println(sum)
		file.WriteString(fmt.Sprintf("%v machines: %v\n", NUM_WORKERS, elapsed))
		elapsedavg += float64(elapsed)
	}
	file.WriteString(fmt.Sprintf("AVERAGE %v machines: %v\n", NUM_WORKERS, elapsedavg / 10))
	file.Close()

	// Cleanup RPC servers, free Unix sockets
	blr.EndRPCServerUnix()
	for _, k := range workers {
		k.EndRPCServerUnix()
	}
}