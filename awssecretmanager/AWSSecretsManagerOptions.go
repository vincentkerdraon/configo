package awssecretmanager

import (
	"log/slog"

	"github.com/vincentkerdraon/configo/lock"
)

type (
	Options struct {
		Logger      *slog.Logger
		Cache       Cache
		ImplCacheID string
		Lock        lock.Locker
	}

	OptionsF func(o *Options)
)

// WithLogger to show information about the processing steps
func WithLogger(l *slog.Logger) OptionsF {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithCache adds a caching layer to avoid calling many time the same secret in a row, for example for JSON document secrets.
//
// A cache with TTL is recommended to increase speed and reduce cost.
// See cachelruttl.
//
// Set implCacheID in the case of the same cache used in different implementation. To avoid key collision. Can be empty.
func WithCache(c Cache, implCacheID string) OptionsF {
	return func(o *Options) {
		o.Cache = c
		o.ImplCacheID = implCacheID
	}
}

// WithLock for a lock when changing values
func WithLock(l lock.Locker) OptionsF {
	return func(o *Options) {
		o.Lock = l
	}
}
