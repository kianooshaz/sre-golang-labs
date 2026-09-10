package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Hour)
	ctx2, cancel := context.WithDeadline(ctx, time.Date(2026, 05, 10, 23, 24, 31, 00, time.UTC))
	defer cancel()

	// Simulate slow operation
	go func(ctx context.Context) {

		repository.Select(ctx)

		select {
		case <-time.After(3 * time.Second):
			ch <- "result"

		case <-ctx.Done():
			fmt.Println("goroutine cancelled")
			return
		}
	}(ctx)

	// Wait with timeout
	select {
	case result := <-ch:
		fmt.Println("Got result:", result)
	case <-time.After(2 * time.Second):
		cancel()
		fmt.Println("Timeout!")
	}

	time.Sleep(3 * time.Second)
}
