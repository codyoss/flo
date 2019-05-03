package flo

import (
	"context"
	"fmt"
	"testing"
)

func TestTypeOfStep(t *testing.T) {
	tests := []struct {
		name string
		step Step
		want stepType
	}{
		{"invalid int", 7, invalid},
		{"invalid string", "", invalid},
		{"invalid bool", false, invalid},
		{"invalid chan", make(chan struct{}), invalid},
		{"invalid struct", struct{}{}, invalid},
		{"invalid ptr", &struct{}{}, invalid},
		{"invalid nil", nil, invalid},
		{"invalid float", 7.7, invalid},
		{"invalid empty fn", func() {}, invalid},
		{"invalid fn 1 wrong in parm", func(i int) {}, invalid},
		{"invalid fn no out parm", func(ctx context.Context) {}, invalid},
		{"invalid fn 1 wrong out parm", func() int { return 0 }, invalid},
		{"invalid fn 1 wrong out parm", func() error { return nil }, invalid},
		{"invalid fn 1 wrong out parm correct in", func(ctx context.Context, s string) int { return 0 }, invalid},
		{"invalid fn 1 wrong out parm correct in", func(ctx context.Context, s string) (int, int) { return 0, 0 }, invalid},
		{"invalid fn 1 wrong in parm correct out", func(s, t string) error { return nil }, invalid},
		{"invalid too many in", func(ctx context.Context, i, j int) error { return nil }, invalid},
		{"invalid too many out", func(ctx context.Context) (bool, bool, error) { return false, false, nil }, invalid},
		{"onlyIn", func(ctx context.Context, b bool) error { return nil }, onlyIn},
		{"inOut", func(ctx context.Context, b bool) (bool, error) { return false, nil }, inOut},
		{"onlyOut", func(ctx context.Context) (bool, error) { return false, nil }, onlyOut},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeOfStep(tt.step); got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidateInputChannelFailures(t *testing.T) {
	invalidstepRunner := &stepRunner{sType: invalid, step: 7}
	inOutstepRunner := &stepRunner{sType: inOut, step: inOutFn}
	tests := []struct {
		name string
		inCh interface{}
		sr   *stepRunner
		want error
	}{
		{"invalid step type", make(chan string, 1), invalidstepRunner, errInputChStepType},
		{"invalid inCh type 1", 7, inOutstepRunner, errInputChType},
		{"invalid inCh type 2", "", inOutstepRunner, errInputChType},
		{"invalid inCh type 3", 7.7, inOutstepRunner, errInputChType},
		{"invalid inCh type 4", struct{}{}, inOutstepRunner, errInputChType},
		{"invalid inCh type 5", &struct{}{}, inOutstepRunner, errInputChType},
		{"invalid inCh type 6", false, inOutstepRunner, errInputChType},
		{"invalid chan dir", createSendChan(), inOutstepRunner, errInputChType},
		{"type mismatch", make(chan int, 1), inOutstepRunner, fmt.Errorf(inputChTypeMismatchFmt, "int", "string")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateInputChannel(tt.inCh, tt.sr); got.Error() != tt.want.Error() {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateOutputChannelFailures(t *testing.T) {
	invalidstepRunner := &stepRunner{sType: invalid, step: 7}
	inOutstepRunner := &stepRunner{sType: inOut, step: inOutFn}
	tests := []struct {
		name string
		inCh interface{}
		sr   *stepRunner
		want error
	}{
		{"invalid step type", make(chan string, 1), invalidstepRunner, errOutputChStepType},
		{"invalid inCh type 1", 7, inOutstepRunner, errOutputChType},
		{"invalid inCh type 2", "", inOutstepRunner, errOutputChType},
		{"invalid inCh type 3", 7.7, inOutstepRunner, errOutputChType},
		{"invalid inCh type 4", struct{}{}, inOutstepRunner, errOutputChType},
		{"invalid inCh type 5", &struct{}{}, inOutstepRunner, errOutputChType},
		{"invalid inCh type 6", false, inOutstepRunner, errOutputChType},
		{"invalid chan dir", createReceiveChan(), inOutstepRunner, errOutputChType},
		{"type mismatch", make(chan int, 1), inOutstepRunner, fmt.Errorf(outputChTypeMismatchFmt, "int", "string")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateOutputChannel(tt.inCh, tt.sr); got.Error() != tt.want.Error() {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithParallelism(t *testing.T) {
	f := NewBuilder()
	if f.parallelism != 1 {
		t.Fatalf("got %d, want 1", f.parallelism)
	}

	f = NewBuilder(WithParallelism(-5))
	if f.parallelism != 1 {
		t.Fatalf("got %d, want 1", f.parallelism)
	}

	f = NewBuilder(WithParallelism(5))
	if f.parallelism != 5 {
		t.Fatalf("got %d, want 5", f.parallelism)
	}
}

func TestWithStepParallelism(t *testing.T) {
	f := NewBuilder()
	if f.parallelism != 1 {
		t.Fatalf("got %d, want 1", f.parallelism)
	}

	f.Add(inOutFn)
	if f.steps[0].parallelism != 1 {
		t.Fatalf("got %d, want 1", f.steps[0].parallelism)
	}

	f = NewBuilder(WithParallelism(5))
	if f.parallelism != 5 {
		t.Fatalf("got %d, want 5", f.parallelism)
	}

	f.Add(inOutFn)
	if f.steps[0].parallelism != 5 {
		t.Fatalf("got %d, want 5", f.steps[0].parallelism)
	}

	f.Add(inOutFn, WithStepParallelism(-5))
	if f.steps[1].parallelism != 5 {
		t.Fatalf("got %d, want 5", f.steps[1].parallelism)
	}

	f.Add(inOutFn, WithStepParallelism(6))
	if f.steps[2].parallelism != 6 {
		t.Fatalf("got %d, want 6", f.steps[2].parallelism)
	}
}

func createSendChan() chan<- bool {
	return make(chan bool)
}

func createReceiveChan() <-chan bool {
	return make(chan bool)
}

func inOutFn(ctx context.Context, s string) (string, error) {
	return "", nil
}
