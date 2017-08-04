package system

import (
	"fmt"
	"net/rpc"
)

// --------------------------------------------------------
// RPC structs

type RegisterArgs struct {
	Worker string
}

type RegisterReply struct {
	Success bool
}

type DoAddWorkArgs struct {
	Balancer string
	Bytes []byte
}

type DoAddWorkReply struct {
	Result int
	Success bool
}

// RPC structs
// --------------------------------------------------------

// --------------------------------------------------------
// Wrapper for dialing servers and calling exported methods on Balancer and Worker

func call(server string, rpcName string, args interface{}, reply interface{}) bool {
	c, errx := rpc.Dial("unix", server)
	if errx != nil {
		return false
	}
	defer c.Close()

	err := c.Call(rpcName, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

// Wrapper for dialing servers and calling exported methods on Balancer and Worker
// --------------------------------------------------------