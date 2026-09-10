package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		// get value metric
		// if value > 10 
		// send alert
		fmt.Printf("Worker %d started job %d\n", id, job)

		time.Sleep(1 * time.Second)

		fmt.Printf("Worker %d finished job %d\n", id, job)
	}
}

func checkMetric() {
	
}

func main() {
	const numberOfWorkers = 3
	const numberOfJobs = 10

	jobs := make(chan int)
	var wg sync.WaitGroup

	// Start workers
	wg.Add(numberOfWorkers)

	for i := 1; i <= numberOfWorkers; i++ {
		go worker(i, jobs, &wg)
	}

	// Send jobs
	go func() {
		for i := 1; i <= numberOfJobs; i++ {
			jobs <- i
		}

		close(jobs)
	}()

	// Wait for all workers
	wg.Wait()

	fmt.Println("All jobs completed")
}
