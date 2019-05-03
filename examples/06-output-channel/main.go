package main

import (
	"context"
	"fmt"

	"github.com/codyoss/flo"
)

func main() {
	inputChannel := make(chan string, 2)
	inputChannel <- "Hello World"
	inputChannel <- "Another message"
	close(inputChannel)
	// make an output channel to receive data from the flo
	outputChannel := make(chan string, 1)

	// Register the output channel
	flo.NewBuilder(flo.WithInput(inputChannel), flo.WithOutput(outputChannel)).
		Add(exclaim).
		Add(exclaim).
		BuildAndExecute(context.Background())

	fmt.Println(<-outputChannel)
	fmt.Println(<-outputChannel)
	// It is your responsibility to close this channel once the flo shutsdown
	close(outputChannel)
	// Output:
	// Hello World!!!
	// Another message!!!
}

func exclaim(ctx context.Context, msg string) (string, error) {
	return msg + "!", nil
}
