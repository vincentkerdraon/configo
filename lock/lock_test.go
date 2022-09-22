package lock

import (
	"context"
	"testing"
	"time"
)

func Test_lock_when_ok(t *testing.T) {
	l := New()
	startLock2Ch := make(chan struct{})
	endLock2Ch := make(chan struct{})
	checkLock2 := 0

	if err := l.LockWithContext(context.Background()); err != nil {
		t.Fatal(err)
	}

	go func() {
		close(startLock2Ch)
		if err := l.LockWithContext(context.Background()); err != nil {
			t.Error(err)
		}
		if checkLock2 != 1 {
			t.Error(checkLock2)
		}
		checkLock2 = 2

		l.Unlock()
		close(endLock2Ch)
	}()

	<-startLock2Ch
	time.Sleep(time.Millisecond)
	checkLock2 = 1

	l.Unlock()
	<-endLock2Ch

	if checkLock2 != 2 {
		t.Error(checkLock2)
	}
}

func Test_lock_when_cancelled(t *testing.T) {
	l := New()

	ctx, cancel := context.WithCancel(context.Background())
	l.LockWithContext(ctx)
	cancel()

	for i := 0; i < 10; i++ {
		if err := l.LockWithContext(ctx); err != context.Canceled {
			t.Fatal(err)
		}
	}
}
