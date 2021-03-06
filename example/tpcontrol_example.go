package main

import (
	"fmt"
	"time"
	"github.com/hekmon/tpcontrol"
)

func main() {
	// Let's go for a throughput of 5 requests by second, 3 differents priorities and and 5 tokens pool size
	flowNbRequests := 5
	flowNbSeconds  := 1
	nbQueues       := 3
	tokenPoolSize  := 5
	scheduler, err := tpcontrol.New(flowNbRequests, flowNbSeconds, nbQueues, tokenPoolSize)
	if err != nil {
		panic(err)
	}

	// Let's wait for the tokenPoolSize to fill up
	requestsEvery := (time.Duration(flowNbSeconds) * time.Second) / time.Duration(flowNbRequests)
	fillUpDuration := requestsEvery * time.Duration(tokenPoolSize)
	fmt.Printf("\nThe token pool size is %d, let's wait %v to let it fill up completly (based on flow defined as %.2f req/s).\n",
					tokenPoolSize, fillUpDuration, float32(flowNbRequests)/float32(flowNbSeconds))
	time.Sleep(fillUpDuration)
	fmt.Println("Time's up !")

	// Spawn priority workers by batches
	notifEnded := make(chan bool)
	launchStarted := time.Now()
	nbBatches := 4
	for currentBatch := 0 ; currentBatch < nbBatches ; currentBatch++ {
		// One per queue (queue 0 == highest priority)
		for currentQueue := 0 ; currentQueue < nbQueues ; currentQueue++ {
			// Be carefull with goroutines and scope
			localBatch := currentBatch
			localQueue := currentQueue
			// Launch worker
			go func() {
				scheduler.CanIGO(localQueue) // This call will block until the scheduler let us work
				fmt.Printf("I am a worker with a priority of %d coming from the batch %d and this experiment started %v ago.\n",
					localQueue, localBatch, time.Since(launchStarted))
				notifEnded <- true // Tell the main goroutine this worker is done
			}()
		}
	}

	// Launchs done
	nbWorkers := nbBatches * nbQueues
	fmt.Printf("\n%d workers launched...\n\n", nbWorkers)

	// Wait for our workers
	nbFinished := 0
	for range notifEnded {
		nbFinished++
		if nbFinished == nbWorkers {
			break
		}
	}

	// Done, thanks for watching
	fmt.Printf("\n%d workers ended their work.\n\n", nbWorkers)

	// Stop the scheduler (not really needed here, but for the example)
	time.Sleep(time.Second)
	scheduler.Stop()
	time.Sleep(time.Second)
}
