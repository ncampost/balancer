package system

import (
	"net"
	"net/rpc"
	"os"
	"io"
	"sync"
	"log"
)



// All the information that defines a Balancer machine
type Balancer struct {
	sync.Mutex
	address string
	workers []string
	chanWorkers chan string
	l net.Listener
}

// --------------------------------------------------------
// Handlers for RPCs that the Balancer receives.

// Workers call this on Balancer when they start up.
func (blr *Balancer) Register(args *RegisterArgs, reply *RegisterReply) error {
	blr.Lock()
	defer blr.Unlock()
	blr.workers = append(blr.workers, args.Worker)
	reply.Success = true
	return nil
}


// Handlers for RPCs that the Balancer receives.
// --------------------------------------------------------

// --------------------------------------------------------
// Internal Balancer functionality

// Initiate an Addition operation on infile, which adds the digits together in infile.
// Expects a file full of digits and no newlines.
func (blr *Balancer) DoAddJob(infile string, chunksize int) int {
	readyWorkers := make(chan string, len(blr.workers))
	quitQueueing := make(chan bool)
	quitAdding := make(chan bool)
	chanResults := make(chan int)
	chanAnswer := make(chan int)
	sum := 0
	var wg sync.WaitGroup

	go blr.handleQueueing(blr.chanWorkers, readyWorkers, quitQueueing)

	for _, k := range blr.workers {
		go func(k string) {
			args := &PingWorkerArgs{blr.address}
			reply := &PingWorkerReply{}
			call(k, "Worker.PingWorker", args, reply)
			if reply.Success {
				readyWorkers <- k
			}
		}(k)
	}

	file, err := os.Open(infile)
	if err != nil {
		log.Fatalf("Error opening %v. Aborting...\n", infile)
	}
	var readerr error = nil
	go func(sum int, chanResults chan int, quitAdding chan bool, chanAnswer chan int) {
		continueAdding := true
		for continueAdding {
			select {
				case result := <- chanResults:
					sum += result
				case <-quitAdding:
					continueAdding = false
					chanAnswer <- sum
			}
		}
	}(sum, chanResults, quitAdding, chanAnswer)
	continueWorking := true
	chanDoneJob := make(chan bool, 1)
	for continueWorking {
		select {
			case targetWorker := <- readyWorkers:
				byteBuffer := make([]byte, chunksize)
				_, readerr = io.ReadAtLeast(file, byteBuffer, chunksize)
				if readerr == io.EOF {
					file.Close()
					chanDoneJob <- true
				} else {
					go func(worker string, byteBuffer []byte, chanResults chan int, chanDoneJob chan bool) {
						wg.Add(1)

						args := &DoAddWorkArgs{blr.address, byteBuffer}
						var reply DoAddWorkReply
						call(targetWorker, "Worker.DoAddWork", args, &reply)

						chanResults <- reply.Result

						blr.chanWorkers <- targetWorker
						wg.Done()
					}(targetWorker, byteBuffer, chanResults, chanDoneJob)
				}
			case <- chanDoneJob:
				continueWorking = false
		}
	}
	wg.Wait()
	quitAdding <- true
	quitQueueing <- true
	answer := <- chanAnswer
	return answer
}

// Initializes new Balancer with provided address.
func MakeBalancer(address string) *Balancer {
	blr := &Balancer{}
	blr.address = address
	blr.workers = []string{}
	blr.chanWorkers = make(chan string)
	blr.StartRPCServerUnix(address)
	return blr
}

// Handles requeuing workers when workers send didWork RPC.
func (blr *Balancer) handleQueueing(chanWorkers chan string, readyWorkers chan string, quitQueueing chan bool) {
	continueQueueing := true
	for continueQueueing {
		select {
			case worker := <- chanWorkers:
				readyWorkers <- worker
			case <- quitQueueing:
				continueQueueing = false
		}
	}
}

// Internal Balancer functionality
// --------------------------------------------------------

// --------------------------------------------------------
// Balancer RPC server

// handles starting and stopping RPC server on provided Unix socket.
func (blr *Balancer) StartRPCServerUnix(socket string) {
	rpcServer := rpc.NewServer()
	rpcServer.Register(blr)
	var e error
	os.Remove(socket)
	blr.l, e = net.Listen("unix", socket)
	
	if e != nil {
		return
	}
	go rpcServer.Accept(blr.l)
}

func (blr *Balancer) EndRPCServerUnix() {
	blr.l.Close()
}

// Balancer RPC server
// --------------------------------------------------------