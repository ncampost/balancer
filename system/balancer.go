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
	address string			 // Name of balancer
	workers []string		 // Names of associated worker machines
	chanWorkers chan string  // Uses channel as synchronized queue to assign jobs to ready workers
	l net.Listener			 // Balancer listens to this object for incoming RPCs
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
	// Initialize relevant channels.
	readyWorkers := make(chan string, len(blr.workers))
	quitQueueing := make(chan bool)
	quitAdding := make(chan bool)
	chanResults := make(chan int)
	chanAnswer := make(chan int)

	
	sum := 0

	// wg makes sure all worker jobs complete before reporting sum.
	var wg sync.WaitGroup

	// Initiate goroutine to handle sending doWork RPCs to ready workers.
	go blr.handleQueueing(blr.chanWorkers, readyWorkers, quitQueueing)

	// Balancer pings all its associated workers to check status, places them in readyWorkers queue.
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

	// Infile to add
	file, err := os.Open(infile)
	if err != nil {
		log.Fatalf("Error opening %v. Aborting...\n", infile)
	}
	var readerr error = nil
	
	// Adding goroutine
	// Goroutine for balancer to collect answers from worker chunks into one sum.
	// Receives results on chanResults: adds result to sum.
	// Receives quitAdding signal: quits adding and reports answer on chanAnswer
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

	// Working goroutine
	// This goroutine handles assigning jobs to workers when they report that they have finished previous job.
	// Receives 
	continueWorking := true
	chanDoneJob := make(chan bool, 1)
	for continueWorking {
		select {
			// This case executes when the balancer detects a ready worker targetWorker.
			case targetWorker := <- readyWorkers:

				// Prepare job for the worker
				byteBuffer := make([]byte, chunksize)
				_, readerr = io.ReadAtLeast(file, byteBuffer, chunksize)
				if readerr == io.EOF {
					// If EOF (no more work to do), close file and send stop signal.
					file.Close()
					chanDoneJob <- true
				} else {
					// Assign job to that worker if not EOF
					go func(worker string, byteBuffer []byte, chanResults chan int, chanDoneJob chan bool) {
						// wg: Make sure all jobs complete before exiting DoAddJob (returning answer)
						wg.Add(1)

						args := &DoAddWorkArgs{blr.address, byteBuffer}
						var reply DoAddWorkReply

						// Blocks until job complete
						call(targetWorker, "Worker.DoAddWork", args, &reply)

						// Report result to the result collection goroutine
						chanResults <- reply.Result

						// Put worker back in the worker queue
						blr.chanWorkers <- targetWorker
						wg.Done()
					}(targetWorker, byteBuffer, chanResults, chanDoneJob)
				}
			// Exit this for loop once we've detected we're done with this job.
			case <- chanDoneJob:
				continueWorking = false
		}
	}
	// Blocks until all worker jobs complete
	wg.Wait()

	// Tell the Adding and Queueing goroutines to stop, since we're done.
	quitAdding <- true
	quitQueueing <- true

	// Receive the answer from the Adding goroutine, and return it.
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