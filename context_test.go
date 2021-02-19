package mergectx

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestContext_cancelRoot(t *testing.T) {

	r, rootCancel := ContextWithCancel(context.Background())

	loop := 10
	var wg sync.WaitGroup
	wg.Add(loop)
	var wg2 sync.WaitGroup
	wg2.Add(loop)

	for i := 0; i < loop; i++ {

		cctx, cancel := r.MergeWithCancel(context.Background())

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

		cctx, cancel := r.MergeWithCancel(context.Background())
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

	cctx, ccancel := r.MergeWithCancel(context.Background())
	defer ccancel()

	parentCancel()
	<-cctx.Done()
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

			_, ccl := r.MergeWithCancel(context.Background())
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