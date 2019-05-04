package flo

import (
	"context"
	"sync"
	"testing"
)

func BenchmarkStep(b *testing.B) {
	b.Skip()
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 1)
	sr := stepRunner{
		sType:       inOut,
		inCh:        in,
		outCh:       out,
		wg:          &sync.WaitGroup{},
		step:        addInt,
		parallelism: 1,
	}
	sr.start(context.Background())

	for n := 0; n < b.N; n++ {
		in <- 1
		<-out
	}

}

func addInt(ctx context.Context, i int) (int, error) {
	return i + i, nil
}
