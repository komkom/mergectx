package mergectx

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContext_cancelRoot(t *testing.T) {

	r, rootCancel := ContextWithCancel(context.Background())

	loop := 10
	var wg sync.WaitGroup
	wg.Add(loop)
	var wg2 sync.WaitGroup
	wg2.Add(loop)

	for i := 0; i < loop; i++ {

		cctx, cancel := r.Merge(context.Background())

		go func() {
			wg.Done()
			<-cctx.Done()
			wg2.Done()
			cancel()
		}()
	}

	wg.Wait()
	rootCancel()
	wg2.Wait()
}

func TestContext_cancelChild(t *testing.T) {

	r, cancel := ContextWithCancel(context.Background())
	defer cancel()

	loop := 10
	var wg sync.WaitGroup
	wg.Add(loop)
	var wg2 sync.WaitGroup
	wg2.Add(loop)

	var cancels []context.CancelFunc
	for i := 0; i < loop; i++ {

		cctx, cancel := r.Merge(context.Background())
		cancels = append(cancels, cancel)

		go func() {
			wg.Done()
			<-cctx.Done()
			wg2.Done()
			cancel()
		}()
	}

	wg.Wait()
	for _, cancel := range cancels {
		cancel()
	}
	wg2.Wait()
}

func TestContext_cancelRootParent(t *testing.T) {

	ctx, parentCancel := context.WithCancel(context.Background())
	r, rootCancel := ContextWithCancel(ctx)
	defer rootCancel()

	cctx, ccancel := r.Merge(context.Background())
	defer ccancel()

	parentCancel()
	<-cctx.Done()
}

func TestContext_cleanChildren(t *testing.T) {

	rootCtx, cancel := ContextWithCancel(context.Background())
	defer cancel()

	loop := 10
	var wg sync.WaitGroup
	wg.Add(loop)

	var cancels []func()

	for i := 0; i < loop; i++ {

		ctx, cancel := rootCtx.Merge(context.Background())
		cancels = append(cancels, cancel)

		go func() {

			<-ctx.Done()
			cancel()
			wg.Done()
		}()
	}

	assert.Equal(t, loop, len(rootCtx.children))

	for _, c := range cancels {
		c()
	}
	wg.Wait()

	assert.Equal(t, 0, len(rootCtx.children))
}

var data = []byte(`___some_data_`)

func BenchmarkRootContext(b *testing.B) {

	r, cancel := ContextWithCancel(context.Background())
	defer cancel()
	ch := make(chan []byte)

	go func() {
		var counter int
		for {
			if counter >= b.N {
				close(ch)
				return
			}

			_, ccl := r.Merge(context.Background())
			ch <- data
			ccl()
			counter++
		}
	}()

	var n int
	for data := range ch {
		n += len(data)
	}

	b.SetBytes(int64(n / b.N))
}

func BenchmarkRaw(b *testing.B) {

	ch := make(chan []byte)

	go func() {
		var counter int
		for {
			if counter >= b.N {
				close(ch)
				return
			}
			ch <- data
			counter++
		}
	}()

	var n int
	for data := range ch {
		n += len(data)
	}

	b.SetBytes(int64(n / b.N))
}

func BenchmarkContextWithCancel(b *testing.B) {

	ch := make(chan []byte)

	go func() {
		var counter int
		for {
			if counter >= b.N {
				close(ch)
				return
			}

			_, cancel := context.WithCancel(context.Background())
			ch <- data
			cancel()
			counter++
		}
	}()

	var n int
	for data := range ch {
		n += len(data)
	}

	b.SetBytes(int64(n / b.N))
}

func BenchmarkContextWithParentCancel(b *testing.B) {

	ch := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		var counter int
		for {
			if counter >= b.N {
				close(ch)
				return
			}

			_, cancel := context.WithCancel(ctx)
			ch <- data
			cancel()
			counter++
		}
	}()

	var n int
	for data := range ch {
		n += len(data)
	}

	b.SetBytes(int64(n / b.N))
}

func BenchmarkContextWithTimeout(b *testing.B) {

	ch := make(chan []byte)

	go func() {
		var counter int
		for {
			if counter >= b.N {
				close(ch)
				return
			}

			_, cancel := context.WithTimeout(context.Background(), time.Second)
			ch <- data
			cancel()
			counter++
		}
	}()

	var n int
	for data := range ch {
		n += len(data)
	}

	b.SetBytes(int64(n / b.N))
}

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

func TestMergeCtx_left(t *testing.T) {

	left, leftCancel := context.WithCancel(context.Background())
	right := context.Background()

	ctx, cancel := mergeCtx(left, right)
	defer cancel()

	leftCancel()
	<-ctx.Done()
}

func TestMergeCtx_right(t *testing.T) {

	left := context.Background()
	right, rightCancel := context.WithCancel(context.Background())

	ctx, cancel := mergeCtx(left, right)
	defer cancel()

	rightCancel()
	<-ctx.Done()
}

func TestMergeCtx_child(t *testing.T) {

	left := context.Background()
	right := context.Background()

	ctx, cancel := mergeCtx(left, right)
	cancel()
	<-ctx.Done()
}

func BenchmarkMergeWithExtraGoroutine(b *testing.B) {

	ctx := context.Background()
	ch := make(chan []byte)

	go func() {
		var counter int
		for {
			if counter >= b.N {
				close(ch)
				return
			}

			_, cancel := mergeCtx(ctx, context.Background())
			ch <- data
			cancel()
			counter++
		}
	}()

	var n int
	for data := range ch {
		n += len(data)
	}

	b.SetBytes(int64(n / b.N))
}
