package param

import (
	"context"
	"time"
)

type (
	Loader struct {
		//SynchroFrequency=0 means only at startup
		SynchroFrequency time.Duration

		//Getter is how to fetch data from the source
		Getter GetterFunc

		// OnChanged callback is called when value changes
		OnChanged func()
	}

	loaderOptions func(r *Loader) error
	GetterFunc    func(ctx context.Context) (string, error)
)

// WithSynchroFrequency is how often the value should be refreshed.
//
// default=0 means only at startup
func WithSynchroFrequency(d time.Duration) loaderOptions {
	return func(l *Loader) error {
		l.SynchroFrequency = d
		return nil
	}
}

// WithCallbackOnChanged to get a callback when the value changes
func WithCallbackOnChanged(f func()) loaderOptions {
	return func(l *Loader) error {
		l.OnChanged = f
		return nil
	}
}

// WithLoader uses a function to fetch data in a local file, secret manager ...
func WithLoader(getter GetterFunc, opts ...loaderOptions) paramOption {
	return func(p *Param) error {
		s := Loader{
			Getter: getter,
		}
		for _, opt := range opts {
			if opt == nil {
				continue
			}
			if err := opt(&s); err != nil {
				return err
			}
		}
		p.Loader = s
		return nil
	}
}
