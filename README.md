[![Go Report Card](https://goreportcard.com/badge/github.com/komkom/mergectx)](https://goreportcard.com/report/github.com/komkom/mergectx)

# mergectx

This pkg implements a context which can be used globally in a server. Incoming request contexts can then be merged with this global context.
On cancellation of the global or a request context the merged context gets also canceled.
This is especially useful for canceling log running request before starting a graceful server shutdown.

A more concrete example here is when using grpc streams then such a context can be useful on the grpc server to cancel stream request before shutting down the server.

```golang
rootCtx, rootCancel := mergectx.ContextWithCancel(context.Background())

loop := 10
var wg sync.WaitGroup
wg.Add(loop)

for i := 0; i < loop; i++ {

  // merge a context with the root context
  ctx, cancel := rootCtx.Merge(context.Background())

  go func() {

    <-ctx.Done()
    wg.Done()
    cancel()
  }()
}

// cancelling the root context now also cancels all the merged contexts
rootCancel()
wg.Wait()

fmt.Printf("stopped\n")
```

There is also a simpler approach to the problem 

```golang
func mergeCtx(left, right context.Context) (context.Context, context.CancelFunc) {

  ctx, cancel := context.WithCancel(right)

  go func() {
    select {
      case <-left.Done():
        cancel()
      case <-ctx.Done():
    }
  }()

   return ctx, cancel
}
```

With this implementation cancelling the left context also cancels the merged context.
Since this approach involves creating a goroutine with every merge its a bit slower. The benachmark `BenchmarkMergeWithExtraGoroutine` uses this approach.

# Benchmarks

```
BenchmarkRootContext
BenchmarkRootContext-16                	  975492	      1229 ns/op	  10.58 MB/s
BenchmarkRaw
BenchmarkRaw-16                        	 2664434	       444 ns/op	  29.26 MB/s
BenchmarkContextWithCancel
BenchmarkContextWithCancel-16          	 1467096	       794 ns/op	  16.38 MB/s
BenchmarkContextWithParentCancel
BenchmarkContextWithParentCancel-16    	  983482	      1205 ns/op	  10.79 MB/s
BenchmarkContextWithTimeout
BenchmarkContextWithTimeout-16         	  130044	      9180 ns/op	   1.42 MB/s
BenchmarkMergeWithExtraGoroutine
BenchmarkMergeWithExtraGoroutine-16    	  475257	      2325 ns/op	   5.59 MB/s
```
