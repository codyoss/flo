package main

import (
	"context"
	"fmt"

	"github.com/codyoss/flo"
)

// Any time an error is returned from a pipeline step that error will not be propagated to the next step. It essentially
// stops the flo for that piece of data. The pipeline is still intact to process future data.
func main() {
	inputChannel := make(chan string, 2)
	inputChannel <- "Hello World"
	inputChannel <- "Another Message"
	close(inputChannel)
	es := errorStep(0)

	// Register a global error handler for all steps. This may be overridden at a step level.
	flo.New(flo.WithInput(inputChannel), flo.WithErrorHandler(globalErrorHandler)).
		Add(es.start).
		Add(es.end, flo.WithStepErrorHandler(stepErrorHandler)). // Override the global error handler.
		Start(context.Background())
	// Output:
	// Handler error at a Global level: Oh no, I failed on the first step
	// Handler error at a Step level: Oh no, I failed on the last step
}

// A function that handles an error
func globalErrorHandler(err error) {
	fmt.Printf("Handler error at a Global level: %v\n", err)
}

func stepErrorHandler(err error) {
	fmt.Printf("Handler error at a Step level: %v\n", err)
}

type errorStep int

func (es *errorStep) start(ctx context.Context, msg string) (string, error) {
	if *es == 0 {
		*es++
		return "", fmt.Errorf("Oh no, I failed on the first step")
	}
	return msg, nil
}

func (es *errorStep) end(ctx context.Context, msg string) error {
	return fmt.Errorf("Oh no, I failed on the last step")
}
