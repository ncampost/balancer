# Balancer: a simple example of distributed computation, written in Go.

Balancer uses Unix sockets to simulate a cluster of machines, one of which is the Balancer, and the rest are Workers.

## Functionality: the cluster takes nums.txt, a file containing 30 million digits, and adds them all together.

The job of the Balancer is to monitor the job, split it up into chunks of fixed size, and hand out the chunks to the Worker machines when they are ready to do work. The Worker machines iterate through the chunks, adding together digits, and report the sum to the Balancer when done.

The test simulates separate machines that communicate via RPC messages.

This system is **not** fault tolerant.

### Runtime

Simply iterating through nums.txt to add the digits takes around `1.6-1.7s`.

(Not counting cluster initialization time!) On my machine, using the Balancer system with 5 Worker machines to do the same task takes around `.95-1s`.

`TODO` I will soon run more tests with different number of Worker machines to gather more runtime data.