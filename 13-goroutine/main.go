package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	// Two producers write names into the channel, main consumes them.
	// wg.Add must happen BEFORE `go`, otherwise wg.Wait can pass too early.

	var wg sync.WaitGroup

	name := make(chan string)

	wg.Add(2)

	go FindNameFromDB("kia", name, &wg)
	go FindNameFromDB("kian", name, &wg)

	// close the channel when both producers are done,
	// so the consumer's `range` can finish
	go func() {
		wg.Wait()
		close(name)
	}()

	for n := range name {
		fmt.Println("name:", n)
	}

	fmt.Println("done")
}

func FindNameFromDB(name string, c chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	time.Sleep(100 * time.Millisecond) // simulate a DB query

	c <- name
}
