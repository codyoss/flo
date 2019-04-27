package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codyoss/flo"
)

func main() {
	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	// Giving the app a little time to process some data, then gracefully shutdown by canceling the context.
	go func() {
		time.Sleep(500 * time.Microsecond)
		cancel()
	}()

	err := flo.NewBuilder(). // Create a flo builder
				Add(helloWorld). // Add a Step
				Add(exclaim).    // Add another
				BuildAndExecute(ctx)       // BuildAndExecute processing(this blocks if there is no error)

	// Checking for validation errors
	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// Hello World! (A bunch of times)
}

func helloWorld(ctx context.Context) (string, error) {
	return "Hello World", nil
}

func exclaim(ctx context.Context, s string) error {
	fmt.Printf("%s!\n", s)
	return nil
}
