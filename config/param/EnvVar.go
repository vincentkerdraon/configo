package param

import (
	"fmt"

	"github.com/vincentkerdraon/configo/config/errors"
)

type (
	EnvVar struct {
		//Name set when different from param name.
		Name string
		Use  bool
	}

	envVarOptions func(*EnvVar) error
)

// WithEnvVarName is the command line argument name to use.
//
// default: param name
func WithEnvVarName(name string) envVarOptions {
	return func(v *EnvVar) error {
		if name == "" {
			return errors.ConfigError{Err: fmt.Errorf("mandatory env var name when using the option")}
		}
		v.Name = name
		return nil
	}
}

// WithReadEnvVar to prevent the flag from being available. Require another mean to get the value.
//
// default: true
func WithReadEnvVar(t bool) envVarOptions {
	return func(v *EnvVar) error {
		v.Use = t
		return nil
	}
}

// WithEnvVar defines optional parameters for environment variables.
func WithEnvVar(opts ...envVarOptions) paramOption {
	return func(p *Param) error {
		v := EnvVar{
			Use: true,
		}
		for _, opt := range opts {
			if opt == nil {
				continue
			}
			if err := opt(&v); err != nil {
				return err
			}
		}
		p.EnvVar = v
		return nil
	}
}
