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

## TODO

- [ ] More docs
- [ ] Better Readme
- [ ] A blog post on this repo