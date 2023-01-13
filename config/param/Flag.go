package param

import (
	"fmt"

	"github.com/vincentkerdraon/configo/config/errors"
)

type (
	Flag struct {
		//Name set when different from param name.
		Name string
		Use  bool
	}

	flagOptions func(*Flag) error
)

// WithFlagName is the command line argument name to use.
//
// default: param name
func WithFlagName(name string) flagOptions {
	return func(f *Flag) error {
		if name == "" {
			return errors.ConfigError{Err: fmt.Errorf("mandatory flag name when using the option")}
		}
		f.Name = name
		return nil
	}
}

// WithReadFlag to prevent the flag from being available. Require another mean to get the value.
//
// default: true
func WithReadFlag(t bool) flagOptions {
	return func(f *Flag) error {
		f.Use = t
		return nil
	}
}

// WithFlag defines optional parameters for command line argument.
func WithFlag(opts ...flagOptions) paramOption {
	return func(p *Param) error {
		f := Flag{
			Use: true,
		}
		for _, opt := range opts {
			if opt == nil {
				continue
			}
			if err := opt(&f); err != nil {
				return err
			}
		}
		p.Flag = f
		return nil
	}
}
