package mergectx_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/komkom/mergectx"
)

func ExampleCtx_cancelMergeCtx() {

	rootCtx, rootCancel := mergectx.ContextWithCancel(context.Background())

	loop := 10
	var wg sync.WaitGroup
	wg.Add(loop)

	for i := 0; i < loop; i++ {

		ctx, cancel := rootCtx.MergeWithCancel(context.Background())

		go func() {

			<-ctx.Done()
			wg.Done()
			cancel()
		}()
	}

	rootCancel()
	wg.Wait()

	fmt.Printf("stopped\n")

	// Output: stopped
}

func ExampleCtx_cancelRequestContexts() {

	rootCtx, cancel := mergectx.ContextWithCancel(context.Background())
	defer cancel()

	loop := 10
	var wg sync.WaitGroup
	wg.Add(loop)

	var cancels []func()

	for i := 0; i < loop; i++ {

		ctx, cancel := rootCtx.MergeWithCancel(context.Background())
		cancels = append(cancels, cancel)

		go func() {

			<-ctx.Done()
			wg.Done()
			cancel()
		}()
	}

	for _, c := range cancels {
		c()
	}
	wg.Wait()

	fmt.Printf("stopped\n")

	// Output: stopped
}
