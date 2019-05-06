# flo

A data pipelining framework that auto-magically(reflection) parallelizes your work`flo`w based upon configuration. All
code is runtime validated before processing begins.(Optionally compile time via exposed testing method)

[![GoDoc](https://godoc.org/github.com/codyoss/flo?status.svg)](https://godoc.org/github.com/codyoss/flo)
[![Build Status](https://cloud.drone.io/api/badges/codyoss/flo/status.svg)](https://cloud.drone.io/codyoss/flo)
[![codecov](https://codecov.io/gh/codyoss/flo/branch/master/graph/badge.svg)](https://codecov.io/gh/codyoss/flo)
[![Go Report Card](https://goreportcard.com/badge/github.com/codyoss/flo)](https://goreportcard.com/report/github.com/codyoss/flo)

*NOTE*, this code trades hiding complexity with a whole lot of reflection, thus it is not that fastest library in the
world. I made it more to prove to myself I could(and to learn the `reflect` package better). Would this code been a
whole lot better if I had generics/contracts to work with? Yep. Once they are released officially I plan to come back to
this project and try to keep a similar api and compare performance. I imagine they will both still heavily use
reflection up front for validation, but with generics I would expect I should not have to pay the price at data
processing time.(maybe) Either way, I think the project is still cool and hey maybe you don't need blazing fast code
and would prefer working with a _slick_ api...

## Overview

Package flo is designed to be an abstraction to the common worker pool pattern used in Go. How does it work? By being
decently opinionated and making heavy use of reflection.

The first thing you need to know about designing a flo is what kind of functions/methods it can work with. Flo's can
consist of one of three function signatures:

```text
 1. func (context.Context) (R, error)
 2. func (context.Context, T) (R, error)
 3. func (context.Context, T) error
```

One can only be used as the first step of a flo. It is meant to act as a step the produces data without an input from
anywhere. Two can be used at any position in the flo, although if it is used as the first or last step in the flo
some extra configuration is expected. Three can only be used as the last step of a flo. It is mean to act as a step
that consumes data and does not send it along to anywhere else.

Now lets break down the common parts of the step signatures. They all take in a context as their first parameter.
This the same context that is passed into the flo when BuildAndExecute is called. It is propagated throughout to
support proper context cancellation. The second things all of the signatures have in common is they all return at
least an error.

Important: If an error is returned from a step in the flo no result will not be propagated to the next step, ending
any data processing for that data stream.

Lastly, you see that steps can optionally take in a T and optionally return an R. Theses are meant to be generic
placeholders for concrete types.

Imagine we have have the following steps:

```text
 func step1(ctx context.Context) (int, error) {...}
 func step2(ctx context.Context, i int) (string, error) {...}
 func step3(ctx context.Context, s string) error {...}
```

We could construct and run a flo with these steps by doing:

```golang
 err := flo.NewBuilder().
            Add(step1).
            Add(step2).
            Add(step3).
            BuildAndExecute(context.Background())
```

This would be an example of a valid flo. Notice how the output type of the previous step must match the input of the
proceeding step(the T and R values mentioned above). If the types did not align properly the err that is returned
would not be nil. The flo validates all types before it begins to process any data. A Validate method is also exposed
from the flo.Builder should you want to validate your flo at test time.

For more detailed examples including how to configure a flo's parallelism and how to bridge a flo to other parts of
your codebase I recommend checkout out the examples folder in this repo.

## Examples

### Basic

```go
func main() {
    // Create a cancelable context
    ctx, cancel := context.WithCancel(context.Background())

    // Giving the app a little time to process some data, then gracefully shutdown.
    go func() {
        time.Sleep(500 * time.Microsecond)
        cancel()
    }()

    err := flo.NewBuilder(). // Create a flo builder
                Add(helloWorld). // Add a Step
                Add(exclaim).    // Add another
                BuildAndExecute(ctx)       // BuildAndExecute processing(this blocks if there is no error)

    // Checking for validation errors
    if err != nil {
        log.Fatal(err)
    }
}
```

### More

For more examples see the [examples folder](examples/)

1. [Running a basic flo](examples/01-basic/main.go)
2. [Registering an input channel to a flo](examples/02-input-channel/main.go)
3. [Configuring flo's parallelism](examples/03-configure-parallelism/main.go)
4. [Validating a flo](examples/04-validation/main.go)
5. [Handling errors in a flo](examples/05-error-handling/main.go)
6. [Registering an output channel for the flo](examples/06-output-channel/main.go)

## Benchmarks

Hey, guess what? Reflection is not super fast. Especially when you are using it as much as this library does. I added
a benchmark in `flo_test.go` which shows how fast flo is compared to hand-writing your own worker pools. This
example also shows how flo can reduce your LOC quite a bit. The flo example is 21 LOC vs 68 LOC for hand writing worker
pools. That is a more than 3x reduction in LOC! So what is the performance cost for this you may ask? 5x.
:sob: :panda_face: Benchmarks were ran on my Razer Blade(2017) Laptop.

```bash
$ go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: github.com/codyoss/flo
BenchmarkFlo-8            200000              6577 ns/op             520 B/op         12 allocs/op
BenchmarkNonFlo-8        1000000              1057 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/codyoss/flo  2.462s
```

## Blog post

I wrote a blog post about this repo. [Check it out here!](https://medium.com/@the.cody.oss/reflecting-on-worker-pools-in-go-7f91f05a5f01)