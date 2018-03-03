package system

import (
	"net"
	"net/rpc"
	"strconv"
	"os"
)

// All the information that defines a Worker machine
type Worker struct {
	address string
	l net.Listener
}

// --------------------------------------------------------
// Handlers for RPCs that the Worker receives.

// Iterate through chunk, summing the digits.
func (wr *Worker) DoAddWork(args *DoAddWorkArgs, reply *DoAddWorkReply) error {
	result := 0
	for _, k := range args.Bytes {
		n, _ := strconv.Atoi(string(k))
		result += n
	}
	reply.Result = result
	reply.Success = true
	return nil
}

// Balancer pings workers at the start of a job to see which workers are ready to take on a job.
// Here we just reply true, but other conditions might have us reply false, depending on other factors not considered here.
func (wr *Worker) PingWorker(args *PingWorkerArgs, reply *PingWorkerReply) error {
	reply.Success = true
	return nil
}

// Handlers for RPCs that the Worker receives.
// --------------------------------------------------------

// --------------------------------------------------------
// Internal Worker functionality

// Initialize a new worker
func MakeWorker(name string, balancer string) *Worker {
	wr := &Worker{address: name}
	args := &RegisterArgs{name}
	var reply RegisterReply
	wr.StartRPCServerUnix(name)
	call(balancer, "Balancer.Register", args, &reply)
	return wr
}

// Internal Worker functionality
// --------------------------------------------------------

// --------------------------------------------------------
// Worker RPC server

// Networking stuff
func (wr *Worker) StartRPCServerUnix(socket string) {
	rpcServer := rpc.NewServer()
	rpcServer.Register(wr)
	var e error
	os.Remove(socket)
	wr.l, e = net.Listen("unix", socket)
	
	if e != nil {
		return
	}
	go rpcServer.Accept(wr.l)
}

func (wr *Worker) EndRPCServerUnix() {
	wr.l.Close()
}


// Worker RPC server
// --------------------------------------------------------