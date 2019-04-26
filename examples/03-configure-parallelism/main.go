package main

import (
	"context"
	"fmt"

	"github.com/codyoss/flo"
)

func main() {
	inputChannel := make(chan string, 1)
	inputChannel <- "Hello World"
	close(inputChannel)

	// Set the default parallelism for the workflow to 5. This means each step added will have 5 worker goroutines and
	// write to a channel with a buffer of 5.
	flo.New(flo.WithInput(inputChannel), flo.WithParallelism(5)).
		Add(exclaim).
		// This step will only have 3 worker goroutines and a output channel with a buffer of 3.
		Add(exclaim, flo.WithStepParallelism(3)).
		Add(print).
		Start(context.Background())
	// Output:
	// Hello World!!
}

func exclaim(ctx context.Context, msg string) (string, error) {
	return msg + "!", nil
}

func print(ctx context.Context, msg string) error {
	fmt.Println(msg)
	return nil
}
