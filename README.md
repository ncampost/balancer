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

Note that my processor has only so many cores, so eventually larger number of worker machine processes start to crowd the processor, and begin to weigh down the system more than they're worth. Though, I would guess that having actual separate machines instead of simulating them on my one machine would remove this slowdown effect. Then, the limiting factor would be additional time spent sending RPC messages versus time saved having more machines do work at the same time, but I would guess that would start happening at a *much* higher machine number.



## TODOs

* I will soon run more tests with different chunksizes to gather more runtime data.
* There's **lots** of error handling that should be being done that is not yet done.
* Handle failures of Worker machines

## Process
1. AddJob is initiated on Balancer, with nums.txt provided as the infile.
2. Balancer initializes a number of channels that are used to synchronize goroutines. One is a buffered channel of length len(blr.workers), which functions as the queue of ready workers.
3. Balancer via RPC pings workers in its slice of worker addresses to check initial availability
4. Balancer creates a goroutine to handle putting workers back in the worker queue when they finish a job (this is a separate goroutine so later we can implement not putting the worker back in the queue if the worker becomes unresponsive), and a goroutine to merge results that the workers will report.
5. Balancer begins reading chunks of bytes out of infile, reading off the channel of ready workers, and sending DoAddWork RPCs of those chunks to the workers it obtains.
6. When the job finishes, the result is commuicated to the result merge goroutine via chan and the worker is put back into the queue via the requeue goroutine.
7. When Balancer reaches the end of file, it sends on channels to quit the merging and requeueing goroutines, and reports the final answer, which is the sum of the whole file.  