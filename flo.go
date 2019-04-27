package flo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	errStepCnt             = errors.New("must register at least two steps")
	errFirstStep           = errors.New("first step must have a signature of func(context.Context) (R, error) or func(context.Context, T) (R, error)")
	errInteriorStep        = errors.New("interior step must have a signature of func(context.Context, T) (R, error)")
	errLastStep            = errors.New("last step must have a signature of func(context.Context, T) error")
	errStepType            = errors.New("a Step must be func with one of the following signatures: func(context.Context) (R, error), func(context.Context, T) (R, error), or func(context.Context, T) error")
	errInputChType         = errors.New("a input channel must be of type <-chan T")
	errInputChStepType     = errors.New("a input channel should only be registered when first step has a signature of func(context.Context, T) (R, error)")
	inputChTypeMismatchFmt = "input channels type %s does not match the first steps input type %s"
	typeMismatchFmt        = "Step %d: previous steps output type %s does not match current steps input type %s"
)

// Builder is used to construct a flo(workflow).
type Builder struct {
	inCh        interface{}
	realChan    chan interface{}
	steps       []*stepRunner
	parallelism int
	errHandler  func(error)
}

// Option that configures a Builder.
type Option func(*Builder)

// WithParallelism configures the default number of workers launched for each step.
func WithParallelism(i int) Option {
	return func(b *Builder) {
		if i < 1 {
			i = 1
		}
		b.parallelism = i
	}
}

// WithErrorHandler configures the default error handler for when a Step returns an error. This is useful should you want
// to do any logging/auditing.
func WithErrorHandler(handler ErrorHandler) Option {
	return func(b *Builder) {
		b.errHandler = handler
	}
}

// WithInput configures the input channel that feeds the pipline.
func WithInput(ch interface{}) Option {
	return func(b *Builder) {
		b.inCh = ch
	}
}

// NewBuilder creates a Builder, used to construct a flo. The flo will not begin to process any data until
// BuildAndExecute is called.
func NewBuilder(options ...Option) *Builder {
	f := &Builder{
		parallelism: 1,
	}

	for i := range options {
		options[i](f)
	}

	return f
}

// Add registers a Step with the flo Builder.
func (b *Builder) Add(s Step, options ...StepOption) *Builder {
	sr := &stepRunner{
		step:        s,
		parallelism: b.parallelism,
		wg:          &sync.WaitGroup{},
		errHandler:  b.errHandler,
	}
	for i := range options {
		options[i](sr)
	}
	b.steps = append(b.steps, sr)
	return b
}

// BuildAndExecute the flo. This will validate all steps registered to the pipeline. If validation fails an error is
// returned and no data will be processed. If validation is successful the steps will begin to process data and this
// method will block until the provdied context is canceled or the input channel closed, if one was registered.
func (b *Builder) BuildAndExecute(ctx context.Context) error {
	err := b.Validate()
	if err != nil {
		return err
	}

	// wire up the steps and start worker pools
	for i := range b.steps {
		if i == 0 {
			if b.inCh != nil {
				b.steps[i].registerInput(b.launchInputChannel())
			}
		} else {
			b.steps[i].registerInput(b.steps[i-1].output())
		}
		b.steps[i].start(ctx)
	}

	b.awaitShutdown()
	return nil
}

// Validate makes sure the pipeline can process data. It ensures all registered steps are of the right type and that
// their input and output types line up. This is the methodd that BuildAndExecute calls. It is exposed mainly for
// testing purposes so that end users of this api can find out at compile time if their pipeline is set up correctly.
func (b *Builder) Validate() error {
	stepCnt := len(b.steps)
	if stepCnt < 2 {
		return errStepCnt
	}

	// loop through and validate steps
	var (
		input      reflect.Type
		output     reflect.Type
		prevOutput reflect.Type
		st         stepType
	)
	for i := range b.steps {
		st = typeOfStep(b.steps[i].step)
		// some initial validation
		if st == invalid {
			return errStepType
		} else if i == 0 && st == onlyIn {
			return errFirstStep
		} else if i == stepCnt-1 && st == onlyOut {
			return errLastStep
		} else if 0 < i && i < stepCnt-1 && st != inOut {
			return errInteriorStep
		}

		// set the stepRunner's type
		b.steps[i].sType = st

		// set variables for input/output types
		switch st {
		case onlyOut:
			output = reflect.TypeOf(b.steps[i].step).Out(0)
		case inOut:
			input = reflect.TypeOf(b.steps[i].step).In(1)
			output = reflect.TypeOf(b.steps[i].step).Out(0)
		case onlyIn:
			input = reflect.TypeOf(b.steps[i].step).In(1)
		}

		if i == 0 {
			prevOutput = output
			continue
		}

		// make sure types align
		if input.Kind() == reflect.Interface {
			if !prevOutput.Implements(input) {
				return fmt.Errorf(typeMismatchFmt, i+1, prevOutput, input)
			}
		} else if prevOutput != input {
			return fmt.Errorf(typeMismatchFmt, i+1, prevOutput, input)
		}
		prevOutput = output
	}

	// validate input channel
	if b.inCh != nil {
		err := validateInputChannel(b.inCh, b.steps[0])
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) launchInputChannel() chan interface{} {
	v := reflect.ValueOf(b.inCh)
	realChan := make(chan interface{}, v.Cap())
	b.realChan = realChan
	go func() {
		for {
			x, ok := v.Recv()
			if !ok {
				close(realChan)
				return
			}
			realChan <- x.Interface()
		}
	}()

	return realChan
}

func validateInputChannel(inCh interface{}, sr *stepRunner) error {
	if sr.sType != inOut {
		return errInputChStepType
	}

	// make sure channel input is a channel that is readable
	t := reflect.TypeOf(inCh)
	if t.Kind() != reflect.Chan ||
		(t.Kind() == reflect.Chan && t.ChanDir() == reflect.SendDir) {
		return errInputChType
	}

	// make sure types align
	input := reflect.TypeOf(sr.step).In(1)
	t = t.Elem()
	if input.Kind() == reflect.Interface {
		if !t.Implements(input) {
			return fmt.Errorf(inputChTypeMismatchFmt, t, input)
		}
	} else if t != input {
		return fmt.Errorf(inputChTypeMismatchFmt, t, input)
	}

	return nil
}

// typeOfStep uses reflection to determine what type of function was passed in as a step.
func typeOfStep(s Step) stepType {
	if s == nil {
		return invalid
	}

	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Func {
		return invalid
	}

	if t.NumIn() < 1 || t.NumIn() > 2 ||
		t.NumOut() < 1 || t.NumOut() > 2 ||
		t.NumIn() == 1 && t.NumOut() == 1 {
		return invalid
	}

	if t.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return invalid
	}

	if t.NumOut() == 1 && t.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return invalid
	}

	if t.NumOut() == 2 && t.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return invalid
	}

	if t.NumIn() == 1 && t.NumOut() == 2 {
		return onlyOut
	}

	if t.NumIn() == 2 && t.NumOut() == 1 {
		return onlyIn
	}

	return inOut
}

func (b *Builder) awaitShutdown() {
	for i := range b.steps {
		b.steps[i].awaitShutdown()
	}
}
