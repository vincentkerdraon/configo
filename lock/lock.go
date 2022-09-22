// Package lock is the equivalent of sync.Mutex but with a context.
//
// Prefer standard sync.Mutex if you don't use the context.
package lock

import (
	"context"
	"sync"
)

type (
	Locker interface {
		LockWithContext(ctx context.Context) error
		Unlock()
		Lock()
	}

	// Lock is like a mutex but with ctx.
	Lock struct {
		lockCh chan struct{}
	}
)

// Lock implements Locker
var _ Locker = (*Lock)(nil)

// Lock implements sync.Locker interface (for convenience)
var _ sync.Locker = (*Lock)(nil)

func New() *Lock {
	return &Lock{lockCh: make(chan struct{}, 1)}
}

// Lock will bock other calls until Unlock() is called or the context is cancelled.
// It will forward the context error when cancelled.
func (l *Lock) LockWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case l.lockCh <- struct{}{}:
		return nil
	}
}

// Lock will bock other calls until Unlock() is called.
//
// See LockWithContext()
func (l *Lock) Lock() {
	_ = l.LockWithContext(context.Background())
}

func (l *Lock) Unlock() {
	<-l.lockCh
}
