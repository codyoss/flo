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

    err := flo.New(). // Create a flo builder
                Add(helloWorld). // Add a Step
                Add(exclaim).    // Add another
                Start(ctx)       // Start processing(this blocks if there is no error)

    // Checking for validation errors
    if err != nil {
        log.Fatal(err)
    }
}
```

### More

For more examples see the [examples folder](examples/)

## TODO

- [ ] More docs
- [ ] Better Readme
- [ ] Examples
- [ ] Close if not 100% coverage (With how much reflection this has.... best to be safe)
- [ ] Timeout for step?
- [ ] Maybe handle `ctrl C` with `flo.Option`
- [ ] Benchmarks
- [ ] A blog post on this repo
- [ ] Add output channel if ending in func(ctx, T) R err