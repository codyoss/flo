package flo

import (
	"context"
	"reflect"
	"sync"
)

type stepType int

const (
	invalid = stepType(iota)
	onlyOut
	onlyIn
	inOut
)

// Step should be a function. The func can look like any of the following examples:
//  func(context.Context) (R, error)
//  func(context.Context, T) (R, error)
//  func(context.Context, T) error
//
// Basically, a Step must at least take a context as its first input parameter and return at least an error. The middle
// example may be used at any point in the Flo. The first example may only be used as the first step of a Flo, and the
// last example may only be used as the last step of a Flo.
type Step interface{}

// ErrorHandler is a function that takes an error. It allows the user to do something when an error occurs.
type ErrorHandler func(error)

type processFn func(context.Context)

// stepRunner orchestrates a worker pool of steps.
type stepRunner struct {
	parallelism int
	inCh        chan interface{}
	outCh       chan interface{}
	step        Step
	wg          *sync.WaitGroup
	sType       stepType
	errHandler  func(error)
}

// StepOption configures how a Step will be run.
type StepOption func(*stepRunner)

// WithStepParallelism configures how many goroutines to launch for the steps worker pool.
func WithStepParallelism(parallelism int) StepOption {
	return func(s *stepRunner) {
		if parallelism < 1 {
			return
		}
		s.parallelism = parallelism
	}
}

// WithStepErrorHandler configures a handler for when a Step returns an error. This is useful should you want to do any
// logging/auditing.
func WithStepErrorHandler(handler ErrorHandler) StepOption {
	return func(s *stepRunner) {
		s.errHandler = handler
	}
}

// registerInput is used to tell the worker pool what channel to listen for data on.
func (s *stepRunner) registerInput(in chan interface{}) {
	s.inCh = in
}

// output creates and returns the output channel for the worker pool.
func (s *stepRunner) output() chan interface{} {
	if s.outCh == nil {
		s.outCh = make(chan interface{}, s.parallelism)
	}
	return s.outCh
}

// start launches the pool of workers for the given step.
func (s *stepRunner) start(ctx context.Context) {
	fn := s.determineProcessFn()
	for i := 0; i < s.parallelism; i++ {
		go func() {
			fn(ctx)
			s.wg.Done()
		}()
		s.wg.Add(1)
	}
}

// determineProcessFn figures out which type of processing to do based on the step's type.
func (s *stepRunner) determineProcessFn() processFn {
	var fn processFn
	switch s.sType {
	case onlyOut:
		fn = s.processOnlyOut
	case inOut:
		fn = s.processInOut
	case onlyIn:
		fn = s.processOnlyIn
	}

	return fn
}

// processOnlyOut is a step that emits data. Could only be the first step in the flo.
func (s *stepRunner) processOnlyOut(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			vs := reflect.ValueOf(s.step).Call([]reflect.Value{reflect.ValueOf(ctx)})
			value := vs[0].Interface()
			err, ok := vs[1].Interface().(error)
			if ok && err != nil {
				if s.errHandler != nil {
					s.errHandler(err)
				}
				continue
			}
			s.outCh <- value
		}
	}
}

// processInOut is a step that has both input and output. Could be any step in the flo.
func (s *stepRunner) processInOut(ctx context.Context) {
	for input := range s.inCh {
		vs := reflect.ValueOf(s.step).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(input)})
		value := vs[0].Interface()
		err, ok := vs[1].Interface().(error)
		if ok && err != nil {
			if s.errHandler != nil {
				s.errHandler(err)
			}
			continue
		}
		s.outCh <- value
	}
}

// processOnlyIn is a step that consumes data. Could only be the last step in the flo.
func (s *stepRunner) processOnlyIn(ctx context.Context) {
	for input := range s.inCh {
		vs := reflect.ValueOf(s.step).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(input)})
		err, ok := vs[0].Interface().(error)
		if ok && err != nil {
			if s.errHandler != nil {
				s.errHandler(err)
			}
			continue
		}
	}
}

// awaitShutdown gracefully shuts down the pool of workers.
func (s *stepRunner) awaitShutdown() {
	s.wg.Wait()
	if s.outCh != nil {
		close(s.outCh)
	}
}
