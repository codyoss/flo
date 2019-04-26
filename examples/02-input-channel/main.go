package main

import (
	"context"
	"fmt"

	"github.com/codyoss/flo"
)

func main() {
	// make an input channel to feed the pipeline some data
	inputChannel := make(chan string, 2)
	// put some data on the channel
	inputChannel <- "Hello World"
	inputChannel <- "Another message"
	// Closing the channel will trigger the pipeline to shutdown. Regularly you would be keep putting data on the
	// channel and close in another goroutine(and not before the pipline was created), but we are trying to keep things
	// simple here.
	close(inputChannel)

	flo.New(flo.WithInput(inputChannel)).
		Add(exclaim).
		Add(exclaim).
		Add(exclaim).
		Add(print).
		Start(context.Background())
	// Output:
	// Hello World!!!
	// Another message!!!
}

func exclaim(ctx context.Context, msg string) (string, error) {
	return msg + "!", nil
}

func print(ctx context.Context, msg string) error {
	fmt.Println(msg)
	return nil
}
