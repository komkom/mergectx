package mergectx

import (
	"context"
	"sync"
)

type Ctx struct {
	context.Context
	mtx      sync.Mutex
	children map[*context.Context]context.CancelFunc
}

func Background() *Ctx {
	return &Ctx{Context: context.Background(), children: make(map[*context.Context]context.CancelFunc)}
}

func Context(ctx context.Context) *Ctx {
	return newCtx(ctx)
}

func ContextWithCancel(ctx context.Context) (*Ctx, context.CancelFunc) {

	ctx, cancel := context.WithCancel(ctx)
	r := newCtx(ctx)
	return r, cancel
}

func newCtx(ctx context.Context) *Ctx {

	r := &Ctx{Context: ctx, children: make(map[*context.Context]context.CancelFunc)}
	go func() {

		<-r.Done()

		r.mtx.Lock()
		for _, cancel := range r.children {
			cancel()
		}
		r.children = nil
		r.mtx.Unlock()
	}()
	return r
}

func (r *Ctx) MergeWithCancel(ctx context.Context) (context.Context, context.CancelFunc) {

	out, cancel := context.WithCancel(ctx)

	r.mtx.Lock()
	r.children[&ctx] = cancel
	r.mtx.Unlock()

	return out, func() {

		r.mtx.Lock()
		delete(r.children, &ctx)
		r.mtx.Unlock()
		cancel()
	}
}
