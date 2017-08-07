# Balancer: a simple example of distributed computation, written in Go.

Balancer uses Unix sockets to simulate a cluster of machines, one of which is the Balancer, and the rest are Workers.

## Functionality: the cluster takes nums.txt, a file containing 30 million digits, and adds them all together.

The job of the Balancer is to monitor the job, split it up into chunks of fixed size, and hand out the chunks to the Worker machines when they are ready to do work. The Worker machines iterate through the chunks, adding together digits, and report the sum to the Balancer when done. The Balancer handles merging Worker results into a final sum of about 135 million.

The test simulates separate machines that communicate via RPC messages.

This system is **not** fault tolerant.

### Runtime

*These are numbers coming from running on my machine (the processor is Intel(R) Core(TM) i5-6200U CPU @ 2.30GHz- it has two cores- and I'm running Ubuntu 16.10. I'm **not** a hardware or OS expert, but I assume that this OS handles switching across and utilizing both cores nicely, so having a nicer processor- particularly, having more cores- would result in greater speedup).*

Simply iterating through nums.txt to add the digits takes around `1.6-1.7s`.

Runtimes averaged over 10 trials. Chunksize for all these trials is 30,000 bytes.

Number of worker machines | Average runtime, 10 trials
--------------------------|---------------------------
1 | 1.9665s
2 | 1.2939s
3 | 1.0297s
4 | 0.9610s
5 | 0.9524s
6 | 0.9633s
7 | 0.9786s
8 | 0.9932s
9 | 1.0020s
10 | 1.0023s

Note the overhead the Balancer system incurs- 1 machine runtime is on average slower than simply iterating through the file. This is from work the Balancer does splitting up the file, adding together results, etc., and also time spent sending RPC messages.

Note that my processor has only so many cores, so eventually larger number of worker machine processes start to crowd the processor, and begin to weigh down the system more than they're worth. Though, I would guess that having actual separate machines instead of simulating them on my one machine would remove this slowdown effect.



## TODOs

* I will soon run more tests with different chunksizes to gather more runtime data.
* There's **lots** of error handling that should be being done that is not yet done.
* Handle failures of Worker machines