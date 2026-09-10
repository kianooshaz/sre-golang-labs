package main

import (
	"fmt"
	"sync"
)

// resouse counter = 0 in memory
// g1 -> 0 -> +1 -> 1
// g2 -> 1 -> +1 -> 2
// g3 -> 2 -> +1 -> 3
// g4 -> 2 -> +1 -> 3
// g5 -> 3 -> +1 -> 4
//
//
//

func main() {
	var wg sync.WaitGroup
	var mt sync.Mutex

	counter := 0

	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			mt.Lock()
			counter = counter + 1
			mt.Unlock()
		}()
	}

	wg.Wait()

	fmt.Println("Counter:", counter)
}
