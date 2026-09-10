package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {

	messages := make(chan string, 7)

	var wg sync.WaitGroup

	wg.Add(7)

	go func() {
		defer wg.Done()
		messages <- "ping 1"
		fmt.Println("ping 1 send to channel")
	}()

	go func() {
		defer wg.Done()
		messages <- "ping 2"
		fmt.Println("ping 2 send to channel")

	}()

	go func() {
		defer wg.Done()
		messages <- "ping 3"
		fmt.Println("ping 3 send to channel")

	}()

	go func() {
		defer wg.Done()
		messages <- "ping 4"
		fmt.Println("ping 4 send to channel")

	}()

	go func() {
		defer wg.Done()
		messages <- "ping 5"
		fmt.Println("ping 5 send to channel")

	}()

	go func() {
		defer wg.Done()
		messages <- "ping 6"
		fmt.Println("ping 6 send to channel")

	}()

	go func() {
		defer wg.Done()
		messages <- "ping 7"
		fmt.Println("ping 7 send to channel")
	}()

	go func() {
		wg.Wait()
		close(messages)
	}()

	for message := range messages {
		// slow proccess
		time.Sleep(3 * time.Second)
		fmt.Println(message)
	}

}
