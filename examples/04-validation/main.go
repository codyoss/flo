package main

import (
	"context"
	"fmt"

	"github.com/codyoss/flo"
)

func main() {
	// It is a good idea to construct your flo in a method. This way you can, at test time, validate that your flo will
	// construct properly.
	fb := FloBuilder()
	fb.Start(context.Background())
	// Output:
	// Hello World!!
}

// FloBuilder constucts the workflow.
func FloBuilder() *flo.Flo {
	inputChannel := make(chan string, 1)
	inputChannel <- "Hello World"
	close(inputChannel)

	return flo.New(flo.WithInput(inputChannel)).
		Add(exclaim).
		Add(print)
}

func exclaim(ctx context.Context, msg string) (string, error) {
	return msg + "!", nil
}

func print(ctx context.Context, msg string) error {
	fmt.Println(msg)
	return nil
}
