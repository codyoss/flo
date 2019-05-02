package flo_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codyoss/flo"
)

func TestFloValidateFailures(t *testing.T) {
	tests := []struct {
		name string
		b    *flo.Builder
		s1   flo.Step
		s2   flo.Step
		s3   flo.Step
	}{
		{"not enough steps", flo.NewBuilder(), badFunc, nil, nil},
		{"bad step type 1", flo.NewBuilder(), "", "", nil},
		{"bad step type 2", flo.NewBuilder(), 0, 0, nil},
		{"bad step type 3", flo.NewBuilder(), false, false, nil},
		{"bad step type 4", flo.NewBuilder(), badFunc, badFunc, nil},
		{"type mismatch", flo.NewBuilder(), start, square, nil},
		{"type mismatch with interface", flo.NewBuilder(), start, read, nil},
		{"wrong first step", flo.NewBuilder(), end, start, nil},
		{"wrong interior step", flo.NewBuilder(), start, start, end},
		{"wrong last step", flo.NewBuilder(), start, middle, start},
		{"wrong input chan type", flo.NewBuilder(flo.WithInput("")), middle, end, nil},
		{"wrong input chan type", flo.NewBuilder(flo.WithInput(0)), middle, end, nil},
		{"wrong input chan type", flo.NewBuilder(flo.WithInput(false)), middle, end, nil},
		{"input chan type mismatch", flo.NewBuilder(flo.WithInput(make(chan int, 1))), stringThingMiddle, stringThingEnd, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.s1 != nil {
				tt.b.Add(tt.s1)
			}
			if tt.s2 != nil {
				tt.b.Add(tt.s2)
			}
			if tt.s3 != nil {
				tt.b.Add(tt.s3)
			}
			err := tt.b.Validate()
			if err == nil {
				t.Errorf("got nil, want error")
			}
		})
	}
}

func TestFloBuildAndExecuteReturnsErr(t *testing.T) {
	err := flo.NewBuilder().BuildAndExecute(context.Background())
	if err == nil {
		t.Errorf("got %v, want nil", err)
	}
}

func TestFloValidateAssignable(t *testing.T) {
	err := flo.NewBuilder().Add(stringThingBuildAndExecute).Add(stringThingEnd).Validate()
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

func TestFloValidateAssignableInChan(t *testing.T) {
	err := flo.NewBuilder(flo.WithInput(make(chan stringThing, 1))).Add(stringThingMiddle).Add(stringThingEnd).Validate()
	if err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

func TestWithErrorHandler(t *testing.T) {
	inCh := make(chan string, 1)
	eh := &errHandle{}
	if eh.hasHandled {
		t.Fatal("got true, want false")
	}

	inCh <- "test"
	close(inCh)
	err := flo.NewBuilder(flo.WithInput(inCh), flo.WithErrorHandler(eh.handleError)).
		Add(erroringMiddle).
		Add(end).
		BuildAndExecute(context.Background())
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	if !eh.hasHandled {
		t.Fatal("got false, want true")
	}
}

func TestWithSetErrorHandler(t *testing.T) {
	inCh := make(chan string, 1)
	eh := &errHandle{}
	if eh.hasHandled {
		t.Fatal("got true, want false")
	}

	stepEh := &errHandle{}
	if stepEh.hasHandled {
		t.Fatal("got true, want false")
	}

	inCh <- "test"
	close(inCh)
	err := flo.NewBuilder(flo.WithInput(inCh), flo.WithErrorHandler(eh.handleError)).
		Add(erroringMiddle, flo.WithStepErrorHandler(stepEh.handleError)).
		Add(end).
		BuildAndExecute(context.Background())
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	if eh.hasHandled {
		t.Fatal("got true, want false")
	}
	if !stepEh.hasHandled {
		t.Fatal("got false, want true")
	}
}

func TestFloBuildAndExecute(t *testing.T) {
	c := &completeFlo{i: 2}
	if c.i != 2 {
		t.Fatalf("got %d, want 2", c)
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Duration(1*time.Millisecond), cancel)
	err := flo.NewBuilder().
		Add(c.start).
		Add(c.middle).
		Add(c.end).
		BuildAndExecute(ctx)
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	if c.i != 4 {
		t.Fatalf("got %d, want 4", c.i)
	}
}

// helpers

func badFunc() {}

func start(ctx context.Context) (string, error) {
	return "hey", nil
}

func middle(ctx context.Context, s string) (string, error) {
	return strings.ToUpper(s), nil
}

func end(ctx context.Context, s string) error {
	return nil
}

func square(ctx context.Context, i int) (int, error) {
	return i * i, nil
}

func read(ctx context.Context, r io.Reader) error {
	return nil
}

type stringThing string

func (s stringThing) String() string {
	return "stringThing"
}

func stringThingBuildAndExecute(ctx context.Context) (stringThing, error) {
	return "", nil
}

func stringThingMiddle(ctx context.Context, s fmt.Stringer) (fmt.Stringer, error) {
	return s, nil
}

func stringThingEnd(ctx context.Context, s fmt.Stringer) error {
	return nil
}

type errHandle struct {
	hasHandled bool
}

func (e *errHandle) handleError(err error) {
	e.hasHandled = true
}

func erroringMiddle(ctx context.Context, s string) (string, error) {
	return "", fmt.Errorf("Bad things")
}

type completeFlo struct {
	i    int
	once sync.Once
}

func (c *completeFlo) start(ctx context.Context) (int, error) {
	return c.i, nil
}

func (c *completeFlo) middle(ctx context.Context, i int) (int, error) {
	return i * i, nil
}

func (c *completeFlo) end(ctx context.Context, i int) error {
	c.once.Do(func() {
		c.i = i
	})
	return nil
}
